import React, { useState, useEffect } from 'react';
import { Parser } from 'json2csv';



const TableComponent4 = ({ scrapedData }) => {
    return (
      <div className="overflow-x-auto">
          {/* min-w-screen min-h-screen  */}
        <div className="min-h-screen flex items-center justify-center font-sans overflow-hidden">
          <div className="w-full lg:w-full">
            <div className="bg-white shadow-md rounded my-6">
              <div className="overflow-x-auto">
                <table className="min-w-max w-full table-auto">
                  <thead>
                    <tr className="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                      {Object.keys(scrapedData[0] || {}).map((key, index) => (
                        <th key={index} className="py-3 px-6 text-left">
                          {key}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="text-gray-600 text-sm font-light">
                    {scrapedData.map((item, index) => (
                      <tr
                        key={index}
                        className={`border-b border-gray-200 hover:bg-gray-100 ${
                          index % 2 === 0 ? 'bg-white' : 'bg-gray-50'
                        }`}
                      >
                        {Object.values(item).map((value, subIndex) => (
                          <td key={subIndex} className="py-3 px-6 text-left whitespace-nowrap">
                            <div className="flex items-center">
                              <span className="font-medium">{value}</span>
                            </div>
                          </td>
                        ))}
                        <td className="py-3 px-6 text-center hidden">
                          <span className="bg-green-200 text-green-600 py-1 px-3 rounded-full text-xs">
                            Active
                          </span>
                        </td>
                        <td className="py-3 px-6 text-center hidden">
                          <div className="flex item-center justify-center">
                            <div className="w-4 mr-2 transform hover:text-purple-500 hover:scale-110">
                              <svg
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth="2"
                                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                                />
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth="2"
                                  d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                                />
                              </svg>
                            </div>
                            <div className="w-4 mr-2 transform hover:text-purple-500 hover:scale-110">
                              <svg
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth="2"
                                  d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
                                />
                            </svg>
                            </div>
                            <div className="w-4 mr-2 transform hover:text-purple-500 hover:scale-110">
                              <svg
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth="2"
                                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                />
                              </svg>
                            </div>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  };

  

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

          <iframe className="w-full border border-gray-300 rounded-lg" style={{ height: '400px' }} src={iframeUrl}></iframe>

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

