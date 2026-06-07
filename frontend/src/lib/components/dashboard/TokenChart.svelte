<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Chart, LineController, LineElement, PointElement, LinearScale, CategoryScale, TimeScale, Filler, Tooltip } from 'chart.js';

	// Register required Chart.js components
	Chart.register(LineController, LineElement, PointElement, LinearScale, CategoryScale, TimeScale, Filler, Tooltip);

	let { data = [] }: { data: Array<{ time: string; value: number }> } = $props();

	let canvasEl: HTMLCanvasElement;
	let chart: Chart | null = null;

	function formatDate(isoString: string): string {
		try {
			const d = new Date(isoString);
			return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
		} catch {
			return isoString;
		}
	}

	$effect(() => {
		if (!canvasEl) return;

		const labels = data.map((d) => formatDate(d.time));
		const values = data.map((d) => d.value);

		if (chart) {
			chart.data.labels = labels;
			chart.data.datasets[0].data = values;
			chart.update();
		} else {
			const ctx = canvasEl.getContext('2d');
			if (!ctx) return;

			// Create accent gradient
			const gradient = ctx.createLinearGradient(0, 0, 0, 300);
			gradient.addColorStop(0, 'rgba(16, 185, 129, 0.2)');
			gradient.addColorStop(1, 'rgba(16, 185, 129, 0.0)');

			chart = new Chart(canvasEl, {
				type: 'line',
				data: {
					labels,
					datasets: [
						{
							label: 'Tokens Processed',
							data: values,
							borderColor: '#10b981',
							borderWidth: 2,
							pointBackgroundColor: '#10b981',
							pointBorderColor: '#0f1117',
							pointHoverRadius: 6,
							fill: true,
							backgroundColor: gradient,
							tension: 0.3
						}
					]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					plugins: {
						legend: {
							display: false
						},
						tooltip: {
							mode: 'index',
							intersect: false,
							backgroundColor: '#1a1d2e',
							titleColor: '#e2e8f0',
							bodyColor: '#94a3b8',
							borderColor: '#2e3348',
							borderWidth: 1,
							padding: 10,
							displayColors: false
						}
					},
					scales: {
						x: {
							grid: {
								color: '#2e3348',
								drawTicks: false
							},
							ticks: {
								color: '#94a3b8',
								font: {
									family: 'Inter, sans-serif',
									size: 11
								},
								maxRotation: 0,
								autoSkip: true,
								maxTicksLimit: 8
							}
						},
						y: {
							grid: {
								color: '#2e3348',
								drawTicks: false
							},
							ticks: {
								color: '#94a3b8',
								font: {
									family: 'Inter, sans-serif',
									size: 11
								}
							}
						}
					}
				}
			});
		}
	});

	onDestroy(() => {
		if (chart) {
			chart.destroy();
		}
	});
</script>

<div class="ql-card chart-card">
	<div class="chart-header">
		<h3 class="chart-title">Token Throughput (Last 24h)</h3>
	</div>
	<div class="chart-wrapper">
		<canvas bind:this={canvasEl}></canvas>
	</div>
</div>

<style>
	.chart-card {
		padding: 1.25rem;
		height: 100%;
		display: flex;
		flex-direction: column;
	}

	.chart-header {
		margin-bottom: 1rem;
	}

	.chart-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.chart-wrapper {
		position: relative;
		flex-grow: 1;
		height: 250px;
	}
</style>
