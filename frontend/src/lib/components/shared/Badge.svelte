<script lang="ts">
	let { text = '', variant = 'neutral' }: {
		text: string;
		variant?: 'success' | 'danger' | 'warning' | 'info' | 'neutral';
	} = $props();

	const variantClasses = {
		neutral: 'bg-ql-surface-2 text-ql-text border-ql-border',
		success: 'bg-green-500/10 text-green-400 border-green-500/20',
		danger: 'bg-red-500/10 text-red-400 border-red-500/20',
		warning: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
		info: 'bg-blue-500/10 text-blue-400 border-blue-500/20'
	};
</script>

<span class="inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium leading-4 {variantClasses[variant] || variantClasses.neutral}">
	{text}
</span>
