export function trapFocus(node: HTMLElement) {
	const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
	const previousOverflow = document.body.style.overflow;
	const selector =
		'button:not([disabled]), a[href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
	document.body.style.overflow = 'hidden';
	queueMicrotask(() => (node.querySelector<HTMLElement>(selector) ?? node).focus());

	function onKeydown(event: KeyboardEvent) {
		if (event.key !== 'Tab') return;
		const focusable = [...node.querySelectorAll<HTMLElement>(selector)].filter(
			(element) => element.offsetParent !== null,
		);
		if (focusable.length === 0) {
			event.preventDefault();
			return;
		}
		const first = focusable[0];
		const last = focusable.at(-1)!;
		if (event.shiftKey && document.activeElement === first) {
			event.preventDefault();
			last.focus();
		} else if (!event.shiftKey && document.activeElement === last) {
			event.preventDefault();
			first.focus();
		}
	}

	node.addEventListener('keydown', onKeydown);
	return {
		destroy() {
			node.removeEventListener('keydown', onKeydown);
			document.body.style.overflow = previousOverflow;
			previous?.focus();
		},
	};
}
