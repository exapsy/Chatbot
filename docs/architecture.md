# Architecture

## Bot architecture

Bot is a mono-repo architecture which orchestrates the following structures and implementations.

```
// Acts as the orchestrator between chat, communication interfaces and prompt compiler
bot {

    // Chats is all the chats that currently are open in session
    // If the capacity is full the a the least used chat is removed
    chats {
        chatsCapacity: uint16
        
        // Represents a chat, its unique id, user? and its history
        chat {
            history: hashmap
            historyCapacity: uint16
            length: uint16 // keep track of the hashmap's length without spending processing power (well we spend only on accessing the variable when on hashmap it has to see how big it is so cpu + mem)
        }
    }
    
    // Compiler of the clients' questions
    prompt {
        queue: Queue[PromptQuestion] // this is the philosophy, in Golang it's implemented somewhat different
        workers: []Prompt_worker
        ----------------
        Prompt(string) // sends string to a worker and worker sends answer back asychronously
    }
    
    // Communication interfaces
    // for the outside world to talk with the bot
    interfaces {
        http
        grpc
        daemon
        --------------
        Listen() channel <- string // Listens to all the interfaces and returns any prompt from any of them to be processed
    }
    
    // Bus is the way to send the compiled messages to
    // and other clients are supposed to catch up.
    bus {
        ---------------
        SendPrompt(chatId: uuid, answer: string)
    }
}
```

**Flow (Scenario: user asking something):**

Here's a scenario describing how a User would interact from start to finish and get their response

1. **User**: Asks something in HTTP interface
2. **Bot-HTTP-interface**: records that to the **Interfaces** message queue
3. **Bot-Interfaces**: Listen() function returns from a channel the prompt to the bot
4. **Bot**: Listens to the interfaces' prompt
5. **Bot**: Sends the prompt to the **Bot-Prompt** to compile an answer
6. **Bot-Prompt**: Sends the prompt string to a worker to compile an answer
7. **Bot**: Waits and listens to the **Bot-Prompt**'s compilation of the question
8. **Bot**: Sends the answer from the **Bot-Prompt** to the **Bot-Chat**
9. **Bot-Chat**: records the question and the answer to the history
10. **Bot-Chat**: sends the answer back to the bus
11. **Bot-Bus**: records the answer and sends it from the infrastructure to the actual bus (e.g. Mosquito, Kafka, RabbitMQ etc.)
12. **Bot-interface-T**: (stateful - created routine to wait for an answer) while waiting for the answer from the bus (with a timeout), it will either be send via **Websockets**, Daemon or whatever bidirectional way possible. Obviously HTTP alone (without WS) is not the way.
13. **User's-client (browser, CLI etc.)**: Reads the message from the daemon/Websocket etc.
14. **User**: Reads the answer after user's client rendered it on the screen.

## Send a prompt in a chat

**User** => **HTTP Server** / **Daemon** / **\<interface\>**

**\<interface\>** => look below for pseudocode

```
client on UI sends on a websocket
websocket handles the message
websocket sends the message down to http
http sends down to bot
bot gets the information about the message such as: ChatID, Messenger

if chat does not exist
    create new chat

chatId := get_chat_id()
    
answer := send prompt to a worker // answers asychronously or parallel

send_answer_to_chat(chatId)
chat_sends_answer_to_bus(answer) // Like Kafka, Mosquitto, RabbitMQ

bot reads bus and gets that there's a new response to be send to a chat
```

## Architectural decisions

- The fact that a **Chat** and a **(Communication) Interface** (e.g. GRPC, HTTP) are separated, means that you can switch or have access to the same chat through separate interfaces. That is a knowing architectural decision
- **Mono-Repo** - Efficient for small solutions
- Backend separation of: `Domain / Infrastructure / Interfaces / App`. Each with its own purpose.
    - **Domain**: Business logic
    - **Infrastructure**: Everything mostly technical needed to implement the business logic and bring it to life. Such as a database (e.g. postgres) or a bus (Mosquito, Kafka etc.)
    - **Interfaces**: Communication of the client with the bot
    - **App**: The bot orchestrating the domains and running all the required infrastructure
- **Docker**: a well established way to have continuous integration or at least keep track of your services/images and different ways to interact with the bot while also keeping track of the versioning in a tactile way
- **Makefile**: used to run all the necessary scripts to develop and deploy the application
- **Prompter** having access to the **Chat** by reference is a decision to:
    - Allow the prompter to have **access to the chat's history**
- **Sveltekit**: Why not react, why not svelte, why not a Golang GUI, or a CLI
    1. **Svelte** is almost pure javascript, by which I mean there are no DOM trickeries and virtual DOMs
    2. **Svelte-kit** allows for Server-side rendering
    3. **Svelte** is proven to be faster in benchmarks almost to pure JavaScript when Frameworks like Angular or React (yes I called React a framework I'm aware:P no debate for this README) are the most slowest frameworks, and other like Backbone are just outdated or JQuery being more than outdated but redundant to say at least.
    4. **Svelte** allows you to focus more on the business side of things
    5. **Svelte** allows for a nice UI/UX without too much hassle (compilation of C++ code for Qt Framework, pure JavaScript focusing more on the how you will build the UI and not on how the UI will be etc.)
    6. Other kind of ways such as a CLI or a GUI could access easily the bot through GRPC, Socket Daemon or simple HTTP, but not built yet.
- **Where is GraphQL???**: No time - but mostly it's redundant. GraphQL is great when you want to connect different entities together, not send redundant information, ask for specific queries etc. But for a chatbot simple as this and for **its scope** (interview), there is no need currently for such support.
- **Kafka**??: Why Kafka? Why put a bus in the first place? You've got no time! - Okay, let me explain: A bus allows for
  future scalability in case we want to make a distribution of microservices that may want to listen to the bus and new
  messages on the bus whether they're bot messages or anything.
  - Okay, but why kafka?: That's what's the most efficient for mass messaging. Mosquitto is for smaller quantities and
    simpler apps while I stand for scalability and flexibility, and rabbitmq is not that efficient for mass messaging
    and queue messaging.

### TESTING!

1. Should be done with the `testify/suite` package, you'll encounter how it's being done in some tests already. [Example](../internal/libs/lists/lists_test.go)
2. Use `interfaces` as much as possible \(!\) to make it possible to make everything testable without having to backwards engineer or implement the actual structure which may not be testable to start with (e.g. a postgres database is not testable by itself, you need a postgres instance, so make an interface for the postgres database so we can test other things later down the line). [Example](../internal/bot/interfaces/interfaces.go)
3. **Do not return the `struct`**! Return the `interface`. Wrong `type Server interface{}; type server struct {}; func New() *server { return &server{...} }`. Right `type Server interface{}; type server struct {}; func New() Server { return &server{...} }`. Reasoning being, you always know if you actually really implement the interface or you go off the line + it's weird to return the struct when the interface is right there, it exists for a reason let's use it for good practice. [Example](../internal/bot/interfaces/http_server/server.go)
