/**
 * Main entry point for the Home Agent frontend application
 */

import './app.css';
import 'highlight.js/styles/github-dark.css';
import App from './App.svelte';

const app = new App({
  target: document.getElementById('app')!,
});

export default app;
