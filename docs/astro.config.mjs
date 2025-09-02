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
						{ label: 'How to Learn df', slug: 'guides/df' },
						{ label: 'Getting Started', slug: 'guides/getting-started' },
						{ label: 'Data Binding', slug: 'guides/data-binding' },
						{ label: 'Dependency Injection', slug: 'guides/dependency-injection' },
						{ label: 'Application Lifecycle', slug: 'guides/application-lifecycle' },
						{ label: 'Advanced Features', slug: 'guides/advanced-features' },
					],
				},
				{
					label: 'Reference',
					autogenerate: { directory: 'reference' },
				},
			],
		}),
	],
});
