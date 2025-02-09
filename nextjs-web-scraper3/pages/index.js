import React, { useState, useEffect } from 'react';
import { Parser } from 'json2csv';
import Script from 'next/script';
  
// Navbar Component
const Navbar = () => {
  return (
    <nav className="fixed w-full bg-white z-90">
      <div className="container mx-auto px-6 py-4 flex justify-between items-center">
        <div className="text-3xl font-bold text-blue-800">LiftScrape</div>
        <button className="bg-blue-800 hover:bg-blue-700 text-white px-6 py-2 rounded-lg">
          Get Started
        </button>
      </div>
    </nav>
  );
};

// Hero Section
const HeroSection = () => {
  return (
    <div className="min-h-screen bg-blue-900 flex flex-col justify-center items-center text-center text-white">
      <h1 className="text-6xl font-bold mb-6">Scrape Web Data with Ease</h1>
      <p className="text-lg font-light mb-8">
        Automatically extract content from any website and download it in the format you need.
      </p>
      <button className="bg-yellow-500 hover:bg-yellow-600 text-gray-900 font-semibold px-8 py-4 rounded-lg">
        Get Started Now
      </button>
    </div>
  );
};

// Features Section
const FeaturesSection = () => {
  return (
    <div className="py-20 bg-gray-100 text-center">
      <h2 className="text-4xl font-bold mb-12">Why LiftScrape?</h2>
      <div className="container mx-auto grid grid-cols-1 md:grid-cols-3 gap-8">
        <div className="bg-white p-8 rounded-lg shadow-md">
          <h3 className="text-2xl font-bold mb-4">Fast & Efficient</h3>
          <p className="text-gray-600">
            Scrape large volumes of data in a fraction of the time it would take manually.
          </p>
        </div>
        <div className="bg-white p-8 rounded-lg shadow-md">
          <h3 className="text-2xl font-bold mb-4">Customizable</h3>
          <p className="text-gray-600">
            Choose exactly what data to scrape with intuitive CSS selectors.
          </p>
        </div>
        <div className="bg-white p-8 rounded-lg shadow-md">
          <h3 className="text-2xl font-bold mb-4">Export to CSV</h3>
          <p className="text-gray-600">
            Easily download the scraped data as a CSV file, ready for further processing.
          </p>
        </div>
      </div>
    </div>
  );
};

// Main Table Component
const TableWithPagination = ({ data }) => {
  const [currentPage, setCurrentPage] = useState(1);
  const [rowsPerPage] = useState(5); // Change rows per page
  const totalPages = Math.ceil(data.length / rowsPerPage);
  const currentData = data.slice((currentPage - 1) * rowsPerPage, currentPage * rowsPerPage);

  const handlePageChange = (pageNumber) => setCurrentPage(pageNumber);

  const handleDownloadCSV = () => {
    const parser = new Parser();
    const csv = parser.parse(data);
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'scraped_data.csv';
    a.click();
  };

  return (
    <div className="w-full mt-10 bg-white shadow-md rounded-lg">
      <div className="p-6 border-b flex justify-between">
        <h3 className="text-lg font-semibold">Scraped Data</h3>
        <button
          onClick={handleDownloadCSV}
          className="bg-yellow-500 hover:bg-yellow-600 text-gray-900 font-semibold px-6 py-3 rounded-lg"
        >
          Download CSV
        </button>
      </div>
      <div className="overflow-x-auto">
        <table className="min-w-full table-auto">
          <thead className="bg-gray-200 text-gray-600 text-left text-sm font-medium">
            <tr>
              {Object.keys(data[0] || {}).map((key, index) => (
                <th key={index} className="px-6 py-4">{key}</th>
              ))}
            </tr>
          </thead>
          <tbody className="text-gray-700 text-sm">
            {currentData.map((item, index) => (
              <tr key={index} className="border-b hover:bg-gray-50 transition">
                {Object.values(item).map((value, subIndex) => (
                  <td key={subIndex} className="px-6 py-4">{value}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <Pagination totalPages={totalPages} currentPage={currentPage} handlePageChange={handlePageChange} />
    </div>
  );
};

// Pagination Component
const Pagination = ({ totalPages, currentPage, handlePageChange }) => {
  const pageNumbers = [];

  for (let i = 1; i <= totalPages; i++) {
    pageNumbers.push(i);
  }

  return (
    <nav className="mt-6 flex justify-center space-x-2">
      {pageNumbers.map((number) => (
        <button
          key={number}
          onClick={() => handlePageChange(number)}
          className={`px-4 py-2 rounded-lg text-sm font-medium ${
            number === currentPage
              ? 'bg-blue-500 text-white'
              : 'bg-gray-200 text-gray-600 hover:bg-gray-300'
          } transition`}
        >
          {number}
        </button>
      ))}
    </nav>
  );
};

// Home Component
const Home = () => {
  const [url, setUrl] = useState('');
  const [iframeUrl, setIframeUrl] = useState('');
  const [selectors, setSelectors] = useState([]);
  const [scrapedData, setScrapedData] = useState([]);

  const handleUrlChange = (e) => setUrl(e.target.value);

  const handleLoadContent = () => {
    const encodedUrl = encodeURIComponent(url);
    setIframeUrl(`/api/proxy?url=${encodedUrl}`);
  };

  const addSelector = (selector) => {
    if (!selectors.includes(selector)) {
      setSelectors([...selectors, selector]);
    }
  };

  const removeSelector = (selector, index) => {
    setSelectors(selectors.filter((_, idx) => idx !== index));
  };

  const fetchScrapedData = async () => {
    const fields = selectors.reduce((acc, cur, idx) => {
      if (idx > 0) acc[`field${idx}`] = cur.replace(/\.highlight\b\s?/g, '');
      return acc;
    }, {});

    const response = await fetch('http://localhost:8080/scrape', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url: url, lcaSelector: selectors[0].replace(/\.highlight\b\s?/g, ''), groups: { fields } }),
    });

    console.log("json:", JSON.stringify({ url: url, lcaSelector: selectors[0].replace(/\.highlight\b\s?/g, ''), groups: { fields } }));

    if (response.ok) {
      const data = await response.json();
      console.log(data);
      setScrapedData(data);
    }
  };

  useEffect(() => {
    const handleMessage = (event) => {
      if (typeof event.data === 'string') addSelector(event.data);
    };

    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, [selectors]);

  return (
    <div>
      <Script
  id="adsbygoogle-init"
  strategy="afterInteractive"
  crossOrigin="anonymous"
  src= "https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=ca-pub-9744548568508745"
/>
      {/* <Script async src="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=ca-pub-9744548568508745" crossOrigin="anonymous"></Script> */}
      <Navbar />
      <HeroSection />

      <div className="container mx-auto mt-16 p-8">
        <div className="bg-white shadow-md rounded-lg p-8">
          <h2 className="text-3xl font-semibold mb-6">Load and Scrape Data</h2>
          <input
            className="shadow-sm border border-gray-300 rounded-lg py-3 px-4 w-full text-gray-800 mb-6"
            type="text"
            value={url}
            onChange={handleUrlChange}
            placeholder="Enter a URL"
          />
          <button
            className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-3 rounded-lg mb-8 transition"
            onClick={handleLoadContent}
          >
            Load Content
          </button>

          <iframe className="w-full h-5/6 border border-gray-300 rounded-lg" style={{ height: '400px' }} src={iframeUrl}></iframe>

          <div className="mt-10">
            <h3 className="text-xl font-semibold mb-4">Selected CSS Selectors</h3>
            <ul className="space-y-4">
              {selectors.map((selector, index) => (
                <li key={index} className="flex justify-between items-center bg-gray-100 p-4 rounded-lg shadow-sm">
                  {selector}
                  <button
                    className="bg-red-500 text-white px-4 py-2 rounded-lg hover:bg-red-600 transition"
                    onClick={() => removeSelector(selector, index)}
                  >
                    Remove
                  </button>
                </li>
              ))}
            </ul>
          </div>

          <button
            className="bg-blue-600 text-white px-6 py-3 rounded-lg mt-6 hover:bg-blue-700 transition"
            onClick={fetchScrapedData}
          >
            Fetch Data
          </button>

          {scrapedData != null && scrapedData.length > 0 && <TableWithPagination data={scrapedData} />}

        </div>
      </div>

      <FeaturesSection />
    </div>
  );
};

export default Home;

