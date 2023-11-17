<script lang="ts">
    import {faker} from '@faker-js/faker';
    import {onMount} from 'svelte';
    // Components
    import {AppShell, Avatar} from '@skeletonlabs/skeleton';

    // Types
    interface Person {
        id: number;
        avatar: number;
        name: string;
    }

    interface MessageFeed {
        id: number;
        host: boolean;
        avatar: number;
        name: string;
        timestamp: string;
        message: string;
        color: string;
    }

    let elemChat: HTMLElement;
    const lorem = faker.lorem.paragraph();

    // people
    const people: Person[] = []
    let currentPerson: Person = {
        id: 1,
        avatar: 2,
        name: "exapsy"
    }

    // Messages
    let messageFeed: MessageFeed[] = [];
    let currentMessage = '';

    // For some reason, eslint thinks ScrollBehavior is undefined...
    // eslint-disable-next-line no-undef
    function scrollChatBottom(behavior?: ScrollBehavior): void {
        elemChat.scrollTo({top: elemChat.scrollHeight, behavior});
    }

    function getCurrentTimestamp(): string {
        return new Date().toLocaleString('en-US', {hour: 'numeric', minute: 'numeric', hour12: true});
    }

    function addMessage(): void {
        const newMessage = {
            id: messageFeed.length,
            host: true,
            avatar: 48,
            name: 'Jane',
            timestamp: `Today @ ${getCurrentTimestamp()}`,
            message: currentMessage,
            color: 'variant-soft-primary'
        };
        // Update the message feed
        messageFeed = [...messageFeed, newMessage];
        // Clear prompt
        currentMessage = '';
        // Smooth scroll to bottom
        // Timeout prevents race condition
        setTimeout(() => {
            scrollChatBottom('smooth');
        }, 0);
    }

    function onPromptKeydown(event: KeyboardEvent): void {
        if (['Enter'].includes(event.code)) {
            event.preventDefault();
            addMessage();
        }
    }

    // When DOM mounted, scroll to bottom
    onMount(() => {
        scrollChatBottom();
    });
</script>

<AppShell>
    <!-- Slot: Sandbox -->
    <section class="card">
        <div class="chat w-full h-full grid grid-cols-1 lg:grid-cols-[30%_1fr]">
            <!-- Chat -->
            <div class="grid grid-row-[1fr_auto]">
                <!-- Conversation -->
                <section bind:this={elemChat} class="max-h-[500px] p-4 overflow-y-auto space-y-4">
                    {#each messageFeed as bubble}
                        {#if bubble.host === true}
                            <div class="grid grid-cols-[auto_1fr] gap-2">
                                <Avatar src="https://i.pravatar.cc/?img={bubble.avatar}" width="w-12"/>
                                <div class="card p-4 variant-soft rounded-tl-none space-y-2">
                                    <header class="flex justify-between items-center">
                                        <p class="font-bold">{bubble.name}</p>
                                        <small class="opacity-50">{bubble.timestamp}</small>
                                    </header>
                                    <p>{bubble.message}</p>
                                </div>
                            </div>
                        {:else}
                            <div class="grid grid-cols-[1fr_auto] gap-2">
                                <div class="card p-4 rounded-tr-none space-y-2 {bubble.color}">
                                    <header class="flex justify-between items-center">
                                        <p class="font-bold">{bubble.name}</p>
                                        <small class="opacity-50">{bubble.timestamp}</small>
                                    </header>
                                    <p>{bubble.message}</p>
                                </div>
                                <Avatar src="https://i.pravatar.cc/?img={bubble.avatar}" width="w-12"/>
                            </div>
                        {/if}
                    {/each}
                </section>
                <!-- Prompt -->
                <section class="border-t border-surface-500/30 p-4">
                    <div class="input-group input-group-divider grid-cols-[auto_1fr_auto] rounded-container-token">
                        <button class="input-group-shim">+</button>
                        <textarea
                                bind:value={currentMessage}
                                class="bg-transparent border-0 ring-0"
                                name="prompt"
                                id="prompt"
                                placeholder="Write a message..."
                                rows="1"
                                on:keydown={onPromptKeydown}
                        />
                        <button class={currentMessage ? 'variant-filled-primary' : 'input-group-shim'}
                                on:click={addMessage}>
                            <i class="fa-solid fa-paper-plane"/>
                        </button>
                    </div>
                </section>
            </div>
        </div>
    </section>
</AppShell>