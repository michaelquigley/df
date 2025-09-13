// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	base: process.env.BASE_URL || '/df',
	integrations: [
		starlight({
			title: 'df',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/michaelquigley/df' }],
			sidebar: [
				{
					label: 'Guides',
					items: [
						// Each item here is one entry in the navigation menu.
						{ label: 'dynamic foundation', slug: 'guides/df-framework' },
						{ label: 'dynamic foundation for data', slug: 'guides/dd' },
						{ label: 'dynamic foundation for logging', slug: 'guides/dl' },
						{ label: 'dynamic foundation for applications', slug: 'guides/da' }
					],
				},
			],
		}),
	],
});
