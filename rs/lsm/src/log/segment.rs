
use std::io::{self, Write, Read, Seek, SeekFrom};
use std::os::unix::fs::FileExt;
use std::fs::{File, OpenOptions};
use std::path::{Path, PathBuf};
use byteorder::{WriteBytesExt, LittleEndian};
use crc32fast::{Hasher};

use super::LogEntry;

const SEGMENT_FILE_EXT: &'static str = "log";

// Record format:
// +--------+-----+-------+----------+
// | header | key | value | checksum |
// +--------+-----+-------+----------+
//
// Header:
// +------------+--------------+
// | key_length | value_length |
// +------------+--------------+
//     4 bytes      8 bytes
//
// Checksum:
// +-------+
// | crc32 |
// +-------+
//  4 bytes
//
const HEADER_LENGTH: usize = 12;
const CHECKSUM_LENGTH: usize = 4;

// File format:
// +-------+----------+-...-+----------+
// | magic | record_0 | ... | record_n |
// +-------+----------+-...-+----------+
const FILE_MAGIC: [u8; 2] = [0xff, 0xff];

pub struct Segment {
    file: File,
    
    // Following the example of Kafka, each segment of the log is named for the
    // index of the first record that it contains.
    base_offset: u64,

    // pos keeps track of the offset of the next byte to write
    pos: usize,
}

impl Segment {
    pub fn new<P>(dir: P, base_offset: u64) -> io::Result<Segment>
    where
        P: AsRef<Path>,
    {
        let file_path = {
            let mut path_buf = PathBuf::new();
            path_buf.push(dir);
            path_buf.push(format!("{:020}", base_offset));
            path_buf.set_extension(SEGMENT_FILE_EXT);
            path_buf
        };

        let mut file = open_log_file(&file_path)?;
        file.write_all(&FILE_MAGIC)?;
        file.sync_all()?;

        Ok(Segment {
            file,
            base_offset,
            pos: FILE_MAGIC.len(),
        })
    }

    pub fn open<P>(file_path: P) -> io::Result<Segment>
    where
        P: AsRef<Path>,
    {
        let file = open_log_file(&file_path.as_ref())?;
        let file_name = file_path.as_ref().file_name().unwrap().to_str().unwrap();
        let base_offset = match u64::from_str_radix(file_name, 10) {
            Ok(offset) => offset,
            Err(_) => {
                return Err(io::Error::new(
                    io::ErrorKind::InvalidData,
                    "Segment file name is not a valid log offset"
                ))
            }
        };
        let file_len = file.metadata()?.len();

        validate_segment_file(&mut file)?;

        Ok(Segment {
            file,
            base_offset,
            pos: file_len as usize,
        })
    }

    pub fn append(&mut self, key: &[u8], val: &[u8]) -> io::Result<u64> {
        let key_len = key.len();
        let val_len = val.len();
        let record_len = HEADER_LENGTH + key_len + val_len + CHECKSUM_LENGTH;
        let mut f = io::BufWriter::with_capacity(record_len, &mut self.file);
        
        let mut body = Vec::with_capacity(key_len+val_len);
        body.write(key).unwrap();
        body.write(val).unwrap();

        let mut hasher = Hasher::new();
        hasher.update(&body);
        let checksum = hasher.finalize();
        
        // Write header
        f.write_u32::<LittleEndian>(key_len as u32)?;
        f.write_u64::<LittleEndian>(val_len as u64)?;

        // Write body
        f.write_all(&mut body)?;
        f.write_u32::<LittleEndian>(checksum)?;

        f.flush()?;

        let curr_offset = self.pos;
        self.pos += record_len;

        Ok(curr_offset as u64)
    }

    pub fn get(&self, offset: u64) -> io::Result<Option<LogEntry>> {
        self.file.seek(SeekFrom::Current(0))
    }
}

fn open_log_file(path: &Path) -> io::Result<File> {
    OpenOptions::new()
        .create(true)
        .read(true)
        .append(true)
        .open(path)
}

fn validate_segment_file(f: &mut File) -> io::Result<()> {
    let mut magic_bytes = [0u8; 2];
    let size = f.read_at(&mut magic_bytes, 0)?;
    if size < 2 || magic_bytes != FILE_MAGIC {
        return Err(io::Error::new(
            io::ErrorKind::InvalidData,
            "Segment file does not contain valid magic bytes",
        ));
    }
    
    Ok(())
}

pub fn test_log() {
    println!("Test from log");
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Cursor;

    #[test]
    fn test_write() {
        let mut log = Log{ source: Cursor::new(vec![]), last_offset: 0 };

        log.write("name".as_bytes(), "Andrew".as_bytes()).unwrap();
        assert_eq!(vec![
            4, 0, 0, 0, // Key Len
            6, 0, 0, 0, 0, 0, 0, 0, // Value len
            110, 97, 109, 101, // "name"
            65, 110, 100, 114, 101, 119, // "Andrew"
            136, 168, 0, 34 // CRC32 checksum
        ], log.source.into_inner());
        assert_eq!(26, log.last_offset);
    }
}
