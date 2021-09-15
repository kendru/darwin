use tokio::sync::{mpsc, oneshot};

struct MyActor {
    receiver: mpsc::Receiver<ActorMessage>,
    next_id: u32,
}

enum ActorMessage {
    GetUniqueId { respond_to: oneshot::Sender<u32> },
}

impl MyActor {
    fn new(receiver: mpsc::Receiver<ActorMessage>) -> Self {
        MyActor {
            receiver: receiver,
            next_id: 0,
        }
    }

    fn handle_message(&mut self, msg: ActorMessage) {
        match msg {
            ActorMessage::GetUniqueId { respond_to } => {
                self.next_id += 1;
                let _ = respond_to.send(self.next_id);
            },
        }
    }
}

async fn run_my_actor(mut actor: MyActor) {
    // The loop exits when all senders to `receiver` have been dropped, then
    // `actor` itself gets dropped.
    while let Some(msg) = actor.receiver.recv().await {
        actor.handle_message(msg);
    }
}

#[derive(Clone)]
pub struct MyActorHandle {
    sender: mpsc::Sender<ActorMessage>,
}

impl MyActorHandle {
    pub fn new() -> Self {
        let (sender, receiver) = mpsc::channel(8);
        let actor = MyActor::new(receiver);
        tokio::spawn(run_my_actor(actor));

        Self { sender }
    }

    pub async fn get_unique_id(&self) -> u32 {
        let (send, recv) = oneshot::channel();
        let msg = ActorMessage::GetUniqueId {
            respond_to: send,
        };
        let _ = self.sender.send(msg).await;

        recv.await.expect("Actor task has been killed")
    }
}

#[tokio::main]
async fn main() {
    let a1 = MyActorHandle::new();
    let a2 = MyActorHandle::new();

    println!("Actor 1: {}", a1.get_unique_id().await);
    println!("Actor 2: {}", a2.get_unique_id().await);
    println!("Actor 1: {}", a1.get_unique_id().await);
    println!("Actor 1: {}", a1.get_unique_id().await);
    println!("Actor 2: {}", a2.get_unique_id().await);
}
