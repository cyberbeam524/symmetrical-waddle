import React, { useState } from 'react';

const IframeLoader = () => {
    const [url, setUrl] = useState('');

    const handleUrlChange = e => setUrl(e.target.value);
    const handleLoadContent = () => {
        window.open(`/api/proxy?url=${encodeURIComponent(url)}`, '_blank');
    };

    return (
        <div>
            <input type="text" value={url} onChange={handleUrlChange} placeholder="Enter URL" />
            <button onClick={handleLoadContent}>Load Content</button>
        </div>
    );
};

export default IframeLoader;
