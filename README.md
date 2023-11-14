# Bot

## How to start

## Example

## Architecture

### Bot architecture

```
bot {
    chats {
        chatsCapacity: uint16
        
        chat {
            history: hashmap
            historyCapacity: uint16
            length: uint16 // keep track of the hashmap's length without spending processing power (well we spend only on accessing the variable when on hashmap it has to see how big it is so cpu + mem)
        }
    }
    prompt {
        queue: Queue[PromptQuestion] // this is the philosophy, in Golang it's implemented somewhat different
        workers: []Prompt_worker
        ----------------
        Prompt(string) // sends string to a worker and worker sends answer back asychronously
    }
}

Prompt_worker {
    
}
```

### Send a prompt in a chat

**User** => **HTTP Server** / **Daemon** / **\<interface\>**

**\<interface\>** => look below for pseudocode

```
if chat does not exist
    create new chat

chatId := get_chat_id()
    
answer := send prompt to a worker // answers asychronously or parallel

send_answer_to_chat(chatId)
```

## Dev Notes

I have fortunately or unfortunately used for errors camel case and for the rest it's snake case with a capital.
