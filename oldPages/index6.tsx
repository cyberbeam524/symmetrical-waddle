import React, { useState, useEffect } from 'react';
import { Parser } from 'json2csv';

// Navbar Component
const Navbar = () => {
  return (
    <nav className="fixed w-full z-50 bg-transparent">
      <div className="container mx-auto px-6 py-4 flex justify-between items-center">
        <div className="text-2xl font-bold text-gray-800">LiftScrape</div>
        <button className="bg-blue-500 hover:bg-blue-600 text-white px-6 py-2 rounded-full">
          Get Started
        </button>
      </div>
    </nav>
  );
};

// Sidebar Component (Collapsible with History)
const Sidebar = ({ history, onSelectUrl }) => {
  const [isCollapsed, setIsCollapsed] = useState(false);

  const toggleSidebar = () => {
    setIsCollapsed(!isCollapsed);
  };

  return (
    <aside
      className={`fixed top-0 left-0 h-screen bg-gray-100 shadow-lg py-8 px-6 transition-transform duration-300 ${
        isCollapsed ? '-translate-x-full' : 'translate-x-0'
      }`}
      style={{ width: isCollapsed ? '64px' : '256px' }}
    >
      <button
        onClick={toggleSidebar}
        className="absolute top-4 right-[-24px] bg-blue-500 text-white p-2 rounded-full focus:outline-none"
      >
        {isCollapsed ? '>' : '<'}
      </button>

      {!isCollapsed && (
        <div>
          <h2 className="text-lg font-semibold mb-4">History</h2>
          <ul className="space-y-2">
            {history.map((url, index) => (
              <li
                key={index}
                className="text-gray-700 bg-gray-200 p-3 rounded-lg hover:bg-gray-300 cursor-pointer"
                onClick={() => onSelectUrl(url)}
              >
                {url}
              </li>
            ))}
          </ul>
        </div>
      )}
    </aside>
  );
};

// Hero Section
const HeroSection = () => {
  return (
    <section className="min-h-screen flex flex-col justify-center items-center text-center bg-gradient-to-r from-blue-100 to-blue-300 pt-24">
      <h1 className="text-5xl font-bold text-gray-800 mb-6">Scrape Web Data Effortlessly</h1>
      <p className="text-xl font-light text-gray-700 mb-8">
        Easily extract web content and download in a convenient CSV format.
      </p>
      <button className="bg-blue-500 hover:bg-blue-600 text-white px-8 py-3 rounded-full">
        Get Started Now
      </button>
    </section>
  );
};

// Features Section
const FeaturesSection = () => {
  return (
    <section className="py-16 bg-gray-100 text-center">
      <h2 className="text-4xl font-bold mb-12 text-gray-800">Why Use LiftScrape?</h2>
      <div className="container mx-auto grid grid-cols-1 md:grid-cols-3 gap-8">
        <div className="bg-white p-8 rounded-lg shadow-lg">
          <h3 className="text-2xl font-semibold mb-4 text-gray-800">Fast & Efficient</h3>
          <p className="text-gray-600">
            Quickly scrape large volumes of data in no time.
          </p>
        </div>
        <div className="bg-white p-8 rounded-lg shadow-lg">
          <h3 className="text-2xl font-semibold mb-4 text-gray-800">Customizable</h3>
          <p className="text-gray-600">
            Choose exactly what data to extract using our intuitive CSS selector.
          </p>
        </div>
        <div className="bg-white p-8 rounded-lg shadow-lg">
          <h3 className="text-2xl font-semibold mb-4 text-gray-800">Export to CSV</h3>
          <p className="text-gray-600">
            Download your scraped data in a simple CSV format.
          </p>
        </div>
      </div>
    </section>
  );
};

// Table Component with Pagination
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
    <div className="w-full mt-10 bg-white shadow-lg rounded-lg ml-72">
      <div className="p-6 border-b flex justify-between">
        <h3 className="text-lg font-semibold text-gray-800">Scraped Data</h3>
        <button
          onClick={handleDownloadCSV}
          className="bg-blue-500 hover:bg-blue-600 text-white font-bold px-6 py-2 rounded-full"
        >
          Download CSV
        </button>
      </div>
      <div className="overflow-x-auto">
        <table className="min-w-full table-auto">
          <thead className="bg-gray-100 text-gray-600 text-left text-sm font-medium">
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
          className={`px-4 py-2 rounded-full text-sm font-medium ${
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

// Main Home Component
const Home = () => {
  const [url, setUrl] = useState('');
  const [iframeUrl, setIframeUrl] = useState('');
  const [selectors, setSelectors] = useState([]);
  const [scrapedData, setScrapedData] = useState([]);
  const [history, setHistory] = useState([]);

  const handleUrlChange = (e) => setUrl(e.target.value);

  const handleLoadContent = () => {
    const encodedUrl = encodeURIComponent(url);
    setIframeUrl(`/api/proxy?url=${encodedUrl}`);
    setHistory([...history, url]); // Add to history
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

    if (response.ok) {
      const data = await response.json();
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

  const handleSelectUrlFromHistory = (url) => {
    setUrl(url);
    handleLoadContent();
  };

  return (
    <div>
      <Navbar />
      {/* <Sidebar history={history} onSelectUrl={handleSelectUrlFromHistory} /> */}
      <HeroSection />

      <div className="container mx-auto mt-16 p-8 ml-72">
        <div className="bg-white shadow-lg rounded-lg p-8">
          <h2 className="text-2xl font-semibold mb-6 text-gray-800">Load and Scrape Data</h2>
          <input
            className="shadow-sm border border-gray-300 rounded-full py-3 px-4 w-full text-gray-800 mb-6"
            type="text"
            value={url}
            onChange={handleUrlChange}
            placeholder="Enter a URL"
          />
          <button
            className="bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded-full mb-8 transition"
            onClick={handleLoadContent}
          >
            Load Content
          </button>

          <iframe className="w-full border border-gray-300 rounded-lg" style={{ height: '400px' }} src={iframeUrl}></iframe>

          <div className="mt-10">
            <h3 className="text-xl font-semibold mb-4 text-gray-800">Selected CSS Selectors</h3>
            <ul className="space-y-4">
              {selectors.map((selector, index) => (
                <li key={index} className="flex justify-between items-center bg-gray-100 p-4 rounded-lg shadow-sm">
                  {selector}
                  <button
                    className="bg-red-500 text-white px-4 py-2 rounded-full hover:bg-red-600 transition"
                    onClick={() => removeSelector(selector, index)}
                  >
                    Remove
                  </button>
                </li>
              ))}
            </ul>
          </div>

          <button
            className="bg-blue-500 text-white px-6 py-3 rounded-full mt-6 hover:bg-blue-600 transition"
            onClick={fetchScrapedData}
          >
            Fetch Data
          </button>

          {scrapedData.length > 0 && <TableWithPagination data={scrapedData} />}
        </div>
      </div>

      <FeaturesSection />
    </div>
  );
};

export default Home;
