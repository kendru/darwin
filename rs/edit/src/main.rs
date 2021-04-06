use std::{cmp::max, io::{self, BufRead, Read, Write}, panic::catch_unwind, time::Instant};
use std::path::Path;
use std::os::unix::prelude::*;
use term_size;
use termios::Termios;

const EDITOR_VERSION: &str = "0.0.1";
const TAB_SIZE: usize = 8;

#[derive(Debug, PartialEq, Clone, Copy)]
enum EditorKey {
    Character(char),
    Escape,
    ArrowLeft,
    ArrowRight,
    ArrowUp,
    ArrowDown,
    PageUp,
    PageDown,
    Home,
    End,
    Del,
}

#[derive(Clone, Copy)]
struct Cursor {
    x: usize,
    y: usize,
    render_x: usize,
}

struct StatusMessage {
    message: String,
    time: Instant,
}

impl StatusMessage {
    fn new<S: Into<String>>(message: S) -> StatusMessage {
        StatusMessage{
            message: message.into(),
            time: Instant::now(),
        }
    }
}

struct EditorLine {
    source: String,
    render: String,
}

impl EditorLine {
    fn x_to_render_x(&self, x: usize) -> usize {
        let mut render_x = 0;
        for i in 0..x {
            let ch = self.source.chars().nth(i).expect("character offset must be valid");
            if ch == '\t' {
                render_x += TAB_SIZE-1-(render_x % TAB_SIZE);
            }
            render_x += 1;
        }
        return  render_x;
    }
}

struct Editor {
    width: usize,
    height: usize,
    stdin: io::Stdin,
    stdout: io::Stdout,
    buf: String,
    rows: Vec<EditorLine>,
    row_offset: usize,
    col_offset: usize,
    cursor: Cursor,
    filename: String,
    status_message: Option<StatusMessage>,

    termios_orig: Termios,
}

impl Editor {
    fn acquire() -> io::Result<Editor> {
        use termios::*;

        let stdin = io::stdin();
        let stdout = io::stdout();

        let fd = io::stdin().as_raw_fd();
        let mut term  = Termios::from_fd(fd)?;
        let orig = term.clone();

        // Set input/output flags.
        term.c_iflag &= !(BRKINT | ICRNL | INPCK | ISTRIP | IXON);
        term.c_oflag &= !OPOST;
        term.c_cflag |= CS8;
        term.c_lflag &= !(ECHO | ICANON | IEXTEN | ISIG);

        term.c_cc[VMIN] = 0;
        term.c_cc[VTIME] = 1;

        tcsetattr(fd, TCSAFLUSH, &term)?;

        let (width, mut height) = window_size()?;
        height -= 2; // Make room for the status and message bars.

        Ok(Editor{
            width,
            height,
            stdin,
            stdout,
            cursor: Cursor{
                x: 0,
                y: 0,
                render_x: 0,
            },
            buf: String::new(),
            rows: Vec::new(),
            row_offset: 0,
            col_offset: 0,
            filename: "[No Name]".to_string(),
            status_message: None,
            termios_orig: orig,
        })
    }

    fn open(&mut self, filename: &str) -> io::Result<()> {
        let file = std::fs::File::open(filename)?;
        let lines = io::BufReader::new(file).lines();
        for line in lines {
            self.append_line(line?);
        }
        self.filename = Path::new(filename)
            .file_name().expect("filename from file should be valid")
            .to_str().expect("filename should be a valid string")
            .to_string();

        Ok(())
    }

    fn append_line(&mut self, line: String) {
        // Size of render line to allocate is size of original line plus enough
        // to accommodate a full tabstop for every tab character.
        let size = line.chars()
            .filter(|ch| *ch == '\t')
            .fold(line.len(), |acc, _| acc + TAB_SIZE-1);

        let mut render: String = String::with_capacity(size);
        for ch in line.chars() {
            if ch == '\t' {
                render.push(' ');
                while render.len() % TAB_SIZE > 0 {
                    render.push(' ');
                }
            } else {
                render.push(ch);
            }
        }

        self.rows.push(EditorLine{
            render: render,
            source: line,
        });
    }

    fn read_key(&mut self) -> io::Result<EditorKey> {
        let mut cbuf = [0u8; 1];
        self.read_byte_into(&mut cbuf)?;
        let c = cbuf[0] as char;
        if c == '\x1b' {
            let mut esc_seq = [0u8; 3];
            if self.read_byte_into(&mut esc_seq[0..1])? != 1 {
                return Ok(EditorKey::Escape);
            };
            if self.read_byte_into(&mut esc_seq[1..2])? != 1 {
                return Ok(EditorKey::Escape);
            };

            if esc_seq[0] as char == '[' {
                if ('0'..'9').contains(&(esc_seq[1] as char)) {
                    if self.read_byte_into(&mut esc_seq[2..3])? != 1 {
                        return Ok(EditorKey::Escape);
                    };
                    if esc_seq[2] as char == '~' {
                        match esc_seq[1] as char {
                            '1' => return Ok(EditorKey::Home),
                            '3' => return Ok(EditorKey::Del),
                            '4' => return Ok(EditorKey::End),
                            '5' => return Ok(EditorKey::PageUp),
                            '6' => return Ok(EditorKey::PageDown),
                            '7' => return Ok(EditorKey::Home),
                            '8' => return Ok(EditorKey::End),
                            _ => {},
                        }
                    }
                } else {
                    match esc_seq[1] as char {
                        // Arrows
                        'A' => return Ok(EditorKey::ArrowUp),
                        'B' => return Ok(EditorKey::ArrowDown),
                        'C' => return Ok(EditorKey::ArrowRight),
                        'D' => return Ok(EditorKey::ArrowLeft),
                        // Home/End
                        'H' => return Ok(EditorKey::Home),
                        'F' => return Ok(EditorKey::End),
                        _ => {},
                    }
                }
            } else if esc_seq[0] as char == 'O' {
                match esc_seq[1] as char {
                    'H' => return Ok(EditorKey::Home),
                    'F' => return Ok(EditorKey::End),
                    _ => {},
                }
            }

            return Ok(EditorKey::Escape);
        }

        Ok(EditorKey::Character(c))
    }

    fn read_byte_into(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        loop {
            match self.stdin.read(buf) {
                Err(err) if err.kind() == io::ErrorKind::Interrupted => {
                    continue;
                },
                res => return res
            }
        }
    }

    fn process_keypress(&mut self) -> io::Result<EditorAction> {
        let key = self.read_key()?;
        match key {
            EditorKey::ArrowUp |
            EditorKey::ArrowDown |
            EditorKey::ArrowLeft |
            EditorKey::ArrowRight => {
                self.move_cursor(key);
                Ok(EditorAction::Continue)
            }

            EditorKey::PageUp | EditorKey::PageDown => {
                if key == EditorKey::PageUp {
                    self.cursor.y = self.row_offset;
                } else if key == EditorKey::PageDown {
                    self.cursor.y = std::cmp::min(
                        self.row_offset + self.height - 1,
                        self.rows.len(),
                    );
                }

                let dir_key = if key == EditorKey::PageUp {
                    EditorKey::ArrowUp
                } else {
                    EditorKey::ArrowDown
                };

                for _ in 0..self.height {
                    self.move_cursor(dir_key);
                }

                Ok(EditorAction::Continue)
            },

            EditorKey::Home => {
                self.cursor.x = 0;
                Ok(EditorAction::Continue)
            },
            EditorKey::End => {
                if self.cursor.y < self.rows.len() {
                    self.cursor.x = self.row_len(self.cursor.y);
                }
                Ok(EditorAction::Continue)
            },

            EditorKey::Character(c) if c == ctrl_key('q') => Ok(EditorAction::Exit),
            _ => Ok(EditorAction::Continue),
        }
    }

    fn move_cursor(&mut self, key: EditorKey) {
        let mut cur = self.cursor;
        match key {
            EditorKey::ArrowLeft => {
                if cur.x > 0 {
                    cur.x -= 1;
                } else if cur.y > 0 {
                    cur.y -= 1;
                    cur.x = self.row_len(cur.y);
                }
            },
            EditorKey::ArrowRight => {
                if cur.x < self.row_len(cur.y) {
                    cur.x += 1;
                } else if cur.y < self.rows.len() {
                    cur.x = 0;
                    cur.y += 1;
                }
            },
            EditorKey::ArrowUp => if cur.y > 0 {
                cur.y -= 1;
            },
            EditorKey::ArrowDown => if cur.y < self.rows.len() {
                cur.y += 1;
            },
            _ => panic!("only arrow keys should be given to move_cursor")
        };

        let row_len= self.row_len(cur.y);
        if cur.x > row_len {
            cur.x = row_len;
        }

        self.cursor = cur;
    }

    fn row_len(&self, row_idx: usize) -> usize {
        self.rows
            .get(row_idx)
            .map(|row| row.source.len())
            .unwrap_or(0)
    }

    fn scroll(&mut self) {
        // TODO: Move the rendor position update into the cursor itself.
        self.cursor.render_x = if self.cursor.y < self.rows.len() {
            self.rows[self.cursor.y].x_to_render_x(self.cursor.x)
        } else {
            0
        };

        self.row_offset = if self.cursor.y < self.row_offset {
            self.cursor.y
        } else if self.cursor.y >= self.row_offset + self.height {
            self.cursor.y - self.height + 1
        } else {
            self.row_offset
        };

        self.col_offset = if self.cursor.render_x < self.col_offset {
            self.cursor.render_x
        } else if self.cursor.render_x >= self.col_offset + self.width {
            self.cursor.render_x - self.width + 1
        } else {
            self.col_offset
        };
    }

    fn refresh_screen(&mut self) -> io::Result<()> {
        self.scroll();

        self.buf = String::new();

        // Hide the cursor
        self.buf += "\x1b[?25l";

        // Position the cursor at the upper-left corner.
        self.buf += "\x1b[H";

        // Draw each line from the buffer.
        self.draw_rows();
        self.draw_status_bar();
        self.draw_message_bar();

        // Re-position the cursor at the correct position.
        let Cursor{ render_x: cur_x, y: cur_y, .. } = self.cursor;
        self.buf += &format!(
            "\x1b[{};{}H",
            cur_y-self.row_offset+1,
            cur_x-self.col_offset+1,
        );

        // Re-show the cursor.
        self.buf += "\x1b[?25h";

        self.stdout.write_all(self.buf.as_bytes())?;
        self.stdout.flush()?;
        Ok(())
    }

    fn draw_rows(&mut self) {
        let row_count = self.rows.len();
        for y in 0..self.height {
            let file_row = y + self.row_offset;
            if file_row >= row_count {
                if row_count == 0 && y == self.height / 3 {
                    let mut welcome = format!("Text editor -- version {}", EDITOR_VERSION);
                    welcome.truncate(self.width);
                    let mut pad_len = max(0, (self.width - welcome.len()) / 2);
                    if pad_len > 0 {
                        self.buf += "~";
                        pad_len -= 1;
                    }
                    self.buf += " ".repeat(pad_len).as_str();
                    self.buf += welcome.as_str();
                } else {
                    self.buf += "~";
                }
            } else {
                let row: String = self.rows[file_row]
                    .render
                    .chars()
                    .skip(self.col_offset)
                    .take(self.width)
                    .collect();
                self.buf += &row;
            }

            self.buf += "\x1b[K";
            self.buf += "\r\n";
        }
    }

    fn draw_status_bar(&mut self) {
        self.buf += "\x1b[7m"; // Invert display.

        // Left status: filename and line count.
        let lstatus_string = format!(
            "{} - {} lines",
            self.filename.chars().take(20).collect::<String>(),
            self.rows.len(),
        );
        let lstatus_len = std::cmp::min(
            lstatus_string.chars().count(),
            self.width,
        );
        self.buf += &self.trim_to_width(&lstatus_string);

        let rstatus = format!("{}/{}", self.cursor.y+1, self.rows.len());
        let rstatus_len = rstatus.chars().count();

        for i in lstatus_len..self.width {
            if i == self.width - rstatus_len {
                self.buf += &rstatus;
                break;
            }
            self.buf += " ";
        }

        self.buf += "\x1b[m"; // Un-invert display.
        self.buf += "\r\n";
    }

    fn draw_message_bar(&mut self) {
        self.buf += "\x1b[K"; // Clear line?
        if let Some(status_message) = self.status_message.as_ref() {
            if status_message.time.elapsed().as_secs() > 5 {
                self.status_message = None;
                return;
            }
            self.buf += &self.trim_to_width(&status_message.message);
        }
    }

    fn trim_to_width(&self, s: &str) -> String {
        return s.chars().take(self.width).collect();
    }

    pub fn set_status_message<S: Into<String>>(&mut self, message: S) {
        self.status_message = Some(StatusMessage::new(message));
    }
}

impl Drop for Editor {
    fn drop(&mut self) {
        use termios::*;

        let _ = self.stdout.write_all(b"\x1b[2J");
        let _ = self.stdout.write_all(b"\x1b[H");
        let _ = self.stdout.flush();

        tcsetattr(self.stdin.as_raw_fd(), TCSAFLUSH, &self.termios_orig)
            .expect("should restore terminal from raw mode");
    }
}

enum EditorAction {
    Continue,
    Exit,
}

fn window_size() -> io::Result<(usize, usize)> {
    term_size::dimensions()
        .ok_or(io::Error::new(
            io::ErrorKind::Other,
            "could not get terminal dimensions."
        )
    )
}

// Returns the character that is `k` + the `Ctrl` key modifier.
fn ctrl_key(c: char) -> char {
    return ((c as u8) & 0x1f) as char;
}

fn run(args: Vec<String>) -> Result<(), io::Error> {
    // This is created and subsequently dropped in order to enter raw mode on start-up
    // and revert on exit.
    let mut editor = Editor::acquire()?;
    if args.len() > 1 {
        editor.open(&args[1])?;
    }

    editor.set_status_message("HELP: Ctrl-Q to Quit");

    loop {
        editor.refresh_screen()?;
        match editor.process_keypress()? {
            EditorAction::Continue => {},
            EditorAction::Exit => {
                break;
            },
        }
    }

    return Ok(())
}

fn main() {
    std::process::exit(match catch_unwind(|| run(std::env::args().collect())) {
        Ok(Ok(_)) => 0,
        Ok(Err(err)) => {
            println!("Error: {:?}", err);
            1
        },
        Err(_) => {
            println!("Unwound");
            1
        }
    });
}
