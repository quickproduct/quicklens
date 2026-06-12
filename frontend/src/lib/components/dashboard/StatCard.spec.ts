import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import StatCard from './StatCard.svelte';

describe('StatCard', () => {
	it('renders title and value correctly', () => {
		render(StatCard, { title: 'Test Stat', value: '42' });
		
		expect(screen.getByText('Test Stat')).toBeTruthy();
		expect(screen.getByText('42')).toBeTruthy();
	});

	it('renders trend correctly', () => {
		render(StatCard, { title: 'Test Stat', value: '42', trend: '+5%', trendUp: true });
		
		expect(screen.getByText('+5%')).toBeTruthy();
		expect(screen.getByText('vs yesterday')).toBeTruthy();
	});
});
