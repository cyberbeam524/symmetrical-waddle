import React, { useState } from 'react';

const Home = () => {
    const [url, setUrl] = useState('');
    const [iframeUrl, setIframeUrl] = useState('');

    const handleUrlChange = e => setUrl(e.target.value);
    const handleLoadContent = () => {
        setIframeUrl(`/api/proxy?url=${encodeURIComponent(url)}`);
    };

    return (
        <div>
            <h1>Web Content Embedder</h1>
            <input type="text" value={url} onChange={handleUrlChange} placeholder="Enter a URL" />
            <button onClick={handleLoadContent}>Load Content</button>
            <iframe src={iframeUrl} style={{ width: '100%', height: '80vh', border: '1px solid #ccc' }} />
        </div>
    );
};

export default Home;
