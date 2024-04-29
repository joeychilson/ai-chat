<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import useChat, { type ChatRequest } from '$lib/chat.svelte';
	import { marked } from 'marked';

	const chat = useChat();

	let message = '';
	let maxTokens = 100;
	let temperature = 1.0;

	function handleSendMessage() {
		if (message.trim() !== '') {
			const req: ChatRequest = {
				message: message.trim(),
				max_tokens: maxTokens,
				temperature: temperature
			};
			chat.sendMessage(req);
			message = '';
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
	<div class="flex-1 overflow-y-auto p-4 chat-window">
		{#each chat.store.messages as message}
			<div class="mb-4 message">
				<div class="font-bold role">{message.role}:</div>
				<div class="prose ml-4 content">{@html marked(message.content)}</div>
			</div>
		{/each}
	</div>
	<div class="p-4 input-container">
		<div class="flex items-center mb-4 space-x-4">
			<Label for="max-tokens">Max Tokens:</Label>
			<Input
				id="max-tokens"
				type="number"
				bind:value={maxTokens}
				class="w-24 px-2 py-1 text-base"
			/>
			<Label for="temperature">Temperature:</Label>
			<Input
				id="temperature"
				type="number"
				step="0.1"
				bind:value={temperature}
				class="w-24 px-2 py-1 text-base"
			/>
		</div>
		<div class="flex">
			<Input
				type="text"
				bind:value={message}
				on:keypress={handleKeyPress}
				placeholder="Type your message..."
				class="flex-1 px-2 py-1 text-base"
			/>
			<Button class="ml-4 px-4 py-2 text-base" on:click={handleSendMessage}>Send</Button>
		</div>
	</div>
</div>
