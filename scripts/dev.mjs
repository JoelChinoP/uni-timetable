import { spawn } from 'node:child_process';
import { existsSync } from 'node:fs';
import { loadEnvFile } from 'node:process';

// ponytail: Node handles env loading and both processes without runner dependencies.
if (existsSync('backend/.env')) {
	loadEnvFile('backend/.env');
}

const detached = process.platform !== 'win32';
const vite = process.platform === 'win32' ? 'node_modules/.bin/vite.cmd' : 'node_modules/.bin/vite';
const children = [
	spawn(vite, { detached, stdio: 'inherit' }),
	spawn('go', ['-C', 'backend', 'run', './cmd/api'], { detached, stdio: 'inherit' }),
];

let stopping = false;
function stop(signal = 'SIGTERM', exitCode = 0) {
	if (stopping) return;
	stopping = true;
	process.exitCode = exitCode;
	children.forEach((child) => {
		if (!child.pid) return;
		try {
			if (detached) process.kill(-child.pid, signal);
			else child.kill(signal);
		} catch (error) {
			if (error.code !== 'ESRCH') throw error;
		}
	});
}

process.once('SIGINT', () => stop('SIGINT'));
process.once('SIGTERM', () => stop('SIGTERM'));

children.forEach((child) => {
	child.once('error', (error) => {
		console.error(error.message);
		stop('SIGTERM', 1);
	});
	child.once('exit', (code) => {
		if (!stopping) stop('SIGTERM', code ?? 1);
	});
});
