import 'htmx.org';

let isRefreshing = false;

document.addEventListener('htmx:responseError', async function (evt) {
    if (evt.detail.xhr.status === 401 && !isRefreshing) {
        isRefreshing = true;

        try {
            const refreshRes = await fetch('/refresh', {
                method: 'POST',
                credentials: 'include'  
            });

            if (refreshRes.ok) {
                htmx.trigger(evt.detail.elt, 'htmx:trigger', { 
                    delay: 10 
                });
            } else {
                window.location.href = '/login';
            }
        } catch (e) {
            window.location.href = '/login';
        } finally {
            isRefreshing = false;
        }
    }
});
