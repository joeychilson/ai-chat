<script lang="ts">
	import useChat from '$lib/chat.svelte';
	import { marked } from 'marked';

	const chat = useChat();

	let inputMessage = '';

	function handleSendMessage() {
		if (inputMessage.trim() !== '') {
			chat.sendMessage(inputMessage.trim());
			inputMessage = '';
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleSendMessage();
		}
	}
</script>

<svelte:head>
	<title>AI Chat</title>
</svelte:head>

<div class="flex flex-col h-screen">
	{#if chat.store.errors && chat.store.errors.length > 0}
		<div class="p-4 text-red-500 bg-red-100">
			{#each chat.store.errors as error}
				<div>{error.message}</div>
			{/each}
		</div>
	{/if}

	<div class="flex-1 overflow-y-auto p-4 chat-window">
		{#each chat.store.messages as message}
			<div class="mb-4 message">
				<div class="font-bold role">{message.role}:</div>
				<div class="prose ml-4 content">{@html marked(message.content)}</div>
			</div>
		{/each}
	</div>
	<div class="flex p-4 input-container">
		<input
			type="text"
			bind:value={inputMessage}
			on:keypress={handleKeyPress}
			placeholder="Type your message..."
			class="flex-1 px-2 py-1 text-base"
		/>
		<button class="ml-4 px-4 py-2 text-base" on:click={handleSendMessage}>Send</button>
	</div>
</div>
