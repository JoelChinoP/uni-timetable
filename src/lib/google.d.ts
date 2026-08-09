export {};

declare global {
	interface Window {
		google?: {
			accounts: {
				id: {
					initialize(config: {
						client_id: string;
						callback: (response: { credential?: string }) => void;
					}): void;
					renderButton(
						element: HTMLElement,
						options: {
							theme?: string;
							size?: string;
							shape?: string;
							text?: string;
							locale?: string;
						},
					): void;
				};
				oauth2: {
					initTokenClient(config: {
						client_id: string;
						scope: string;
						callback: (response: { access_token?: string; error?: string }) => void;
						error_callback?: () => void;
					}): { requestAccessToken(options?: { prompt?: string }): void };
				};
			};
		};
	}
}
