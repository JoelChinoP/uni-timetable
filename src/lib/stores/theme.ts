import { writable } from 'svelte/store';

export type ThemeMode = 'dark' | 'light';

const STORAGE_KEY = 'uni-timetable-theme';

function getPreferredTheme(): ThemeMode {
	if (typeof window === 'undefined') {
		return 'light';
	}

	const storedValue = window.localStorage.getItem(STORAGE_KEY);
	if (storedValue === 'light' || storedValue === 'dark') {
		return storedValue;
	}

	return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function applyTheme(theme: ThemeMode) {
	if (typeof document === 'undefined') {
		return;
	}

	document.documentElement.dataset.theme = theme;
	document.documentElement.style.colorScheme = theme;
}

function createThemeStore() {
	const initialTheme = getPreferredTheme();
	const { subscribe, set, update } = writable<ThemeMode>(initialTheme);

	applyTheme(initialTheme);

	return {
		subscribe,
		setTheme(theme: ThemeMode) {
			applyTheme(theme);
			if (typeof window !== 'undefined') {
				window.localStorage.setItem(STORAGE_KEY, theme);
			}
			set(theme);
		},
		toggle() {
			update((currentTheme) => {
				const nextTheme = currentTheme === 'dark' ? 'light' : 'dark';
				applyTheme(nextTheme);
				if (typeof window !== 'undefined') {
					window.localStorage.setItem(STORAGE_KEY, nextTheme);
				}
				return nextTheme;
			});
		},
	};
}

export const theme = createThemeStore();
