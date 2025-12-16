/**
 * Main entry point for the Home Agent frontend application
 */

import './app.css';
import 'highlight.js/styles/github-dark.css';
import { mount } from 'svelte';
import App from './App.svelte';

const app = mount(App, {
  target: document.getElementById('app')!,
});

export default app;
