export function extractCsrfToken() {
    if (typeof document !== 'undefined' && typeof document.cookie !== 'undefined') {
        const cookies = document.cookie.split(';');
        for (let i = 0; i < cookies.length; i++) {
            const cookie = cookies[i].trim();
            if (cookie.startsWith('MMCSRF=')) {
                return cookie.replace('MMCSRF=', '');
            }
        }
    }
    return '';
}
