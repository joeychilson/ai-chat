import { SSE, type SSEvent } from 'sse.js';
import { toast } from 'svelte-sonner';

class ChatStore {
	messages: { role: string; content: string }[] = $state([]);
}

export interface ChatRequest {
	message: string;
	max_tokens?: number;
}

export default function useChat() {
	const store = new ChatStore();

	async function sendMessage(req: ChatRequest) {
		store.messages.push({ role: 'user', content: req.message });

		const eventSource = new SSE('/chat', {
			headers: {
				'Content-Type': 'application/json'
			},
			payload: JSON.stringify({ message: req.message, max_tokens: req.max_tokens })
		});

		eventSource.addEventListener('error', () => {
			toast.error('An error occurred while sending the message');
		});

		let currentContentBlock: ContentBlock | undefined;

		eventSource.addEventListener('message', (event: SSEvent) => {
			try {
				const data: StreamEvent = JSON.parse(event.data);

				if (data.type === 'message_start' && data.message) {
					store.messages.push({ role: 'assistant', content: '' });
				} else if (data.type === 'content_block_start' && data.content_block) {
					currentContentBlock = data.content_block;
				} else if (data.type === 'content_block_delta' && currentContentBlock && data.delta) {
					if (currentContentBlock.type === 'text') {
						store.messages[store.messages.length - 1].content += data.delta.text;
					}
				} else if (data.type === 'message_stop') {
					currentContentBlock = undefined;
				}
			} catch (err) {
				console.error('Error during chat:', err);
			}
		});
		eventSource.stream();
	}
	return { store, sendMessage };
}

interface StreamEvent {
	type: string;
	message?: Message;
	content_block?: ContentBlock;
	delta?: {
		text: string;
	};
}

interface Message {
	id: string;
	type: string;
	role: string;
	content: ContentBlock[];
	model: string;
	stop_reason: string | null;
	stop_sequence: string | null;
	usage: {
		input_tokens: number;
		output_tokens: number;
	};
}

interface ContentBlock {
	type: string;
	text: string;
}

interface Error {
	message: string;
}
