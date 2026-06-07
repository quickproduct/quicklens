<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Chart, BarController, BarElement, LinearScale, CategoryScale, Tooltip } from 'chart.js';

	// Register required Chart.js components
	Chart.register(BarController, BarElement, LinearScale, CategoryScale, Tooltip);

	let { data = [] }: { data: Array<{ time: string; value: number }> } = $props();

	let canvasEl: HTMLCanvasElement;
	let chart: Chart | null = null;

	function formatDay(isoString: string): string {
		try {
			const d = new Date(isoString);
			return d.toLocaleDateString([], { weekday: 'short' });
		} catch {
			return isoString;
		}
	}

	$effect(() => {
		if (!canvasEl) return;

		const labels = data.map((d) => formatDay(d.time));
		const values = data.map((d) => d.value);

		if (chart) {
			chart.data.labels = labels;
			chart.data.datasets[0].data = values;
			chart.update();
		} else {
			const ctx = canvasEl.getContext('2d');
			if (!ctx) return;

			// Create bar gradient
			const gradient = ctx.createLinearGradient(0, 0, 0, 250);
			gradient.addColorStop(0, '#3b82f6');
			gradient.addColorStop(1, 'rgba(59, 130, 246, 0.1)');

			chart = new Chart(canvasEl, {
				type: 'bar',
				data: {
					labels,
					datasets: [
						{
							label: 'Total Cost ($)',
							data: values,
							backgroundColor: gradient,
							borderColor: '#3b82f6',
							borderWidth: 1,
							borderRadius: 4,
							borderSkipped: false
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
							backgroundColor: '#1a1d2e',
							titleColor: '#e2e8f0',
							bodyColor: '#94a3b8',
							borderColor: '#2e3348',
							borderWidth: 1,
							padding: 10,
							displayColors: false,
							callbacks: {
								label: function (context) {
									return `Cost: $${Number(context.raw).toFixed(4)}`;
								}
							}
						}
					},
					scales: {
						x: {
							grid: {
								display: false
							},
							ticks: {
								color: '#94a3b8',
								font: {
									family: 'Inter, sans-serif',
									size: 11
								}
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
								},
								callback: function (value) {
									return '$' + Number(value).toFixed(2);
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
		<h3 class="chart-title">Estimated Cost (Last 7 Days)</h3>
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
