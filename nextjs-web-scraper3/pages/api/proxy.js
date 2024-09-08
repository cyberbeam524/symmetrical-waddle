import axios from 'axios';

// Simple in-memory cache
const htmlCache = {};

export default async (req, res) => {
    const { url } = req.query;
    if (!url) {
        return res.status(400).json({ message: 'URL parameter is required' });
    }

        // Return cached content if available
        if (htmlCache[url]) {
            res.setHeader('Content-Type', 'text/html');
            res.send(htmlCache[url]);
            return;
        }

    try {
        const response = await axios.get(url, {
            responseType: 'arraybuffer',
            responseEncoding: 'binary'
        });

        let htmlContent = response.data.toString('utf-8');
        const baseUrl = new URL(url).origin;

        // Modify the HTML content using string replacements
        // Adjust src and href attributes to be absolute URLs
        htmlContent = htmlContent.replace(/(src|href)="([^"]*)"/g, (match, p1, p2) => {
            if (!p2.startsWith('http') && !p2.startsWith('//')) {
                const newUrl = new URL(p2, baseUrl).href;
                return `${p1}="${newUrl}"`;
            }
            return match;
        });

        // Inject custom styles for highlighting elements
        htmlContent = htmlContent.replace('</head>', '<style>.highlight { border: 2px solid green; }</style></head>');

        // Inject JavaScript to enable element highlighting on click
        htmlContent = htmlContent.replace('</body>', `<script>
            document.addEventListener('click', function(event) {
                event.preventDefault();
                const target = event.target;
                target.classList.add('highlight');
                console.log(target.tagName.toLowerCase() + (target.className ? '.' + target.className.split(' ').join('.') : ''));

                const selector = target.tagName.toLowerCase() + (target.className ? '.' + target.className.split(' ').join('.') : '');
                window.parent.postMessage(selector, '*');
            });

            document.addEventListener('DOMContentLoaded', function() {
                window.addEventListener('message', function(event) {
                    // Validate the origin here to ensure messages are from your trusted domain
                    if (event.origin !== "http://localhost:3000") { // Adjust the origin according to your deployment
                        return;
                    }
            
                    const data = event.data;
                    if (data && data.type === 'REMOVE_HIGHLIGHT') {
                        const elements = document.querySelectorAll(data.selector);
                        elements.forEach(el => {
                            el.classList.remove('highlight'); // Assuming 'highlight' is a class that applies the style
                        });
                    }
                });
            });
        </script></body>`);

        res.setHeader('Content-Type', 'text/html');

        htmlCache[url] = htmlContent; // Cache the processed HTML


        res.send(htmlContent);
    } catch (error) {
        console.error('Failed to fetch or process the page:', error);
        res.status(500).json({ message: 'Failed to fetch the page' });
    }
};

