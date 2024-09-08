import React, { useState, useEffect } from 'react';

import { Parser } from 'json2csv';

// Sidebar component
const Sidebar = () => {
    return (
      <div className="h-full w-64 bg-gray-800 text-white fixed">
        <div className="p-4">
          <h2 className="text-2xl font-bold">Sidebar</h2>
          <ul className="mt-4">
            <li className="p-2 hover:bg-gray-700 cursor-pointer">Dashboard</li>
            <li className="p-2 hover:bg-gray-700 cursor-pointer">Scrape Data</li>
            <li className="p-2 hover:bg-gray-700 cursor-pointer">Settings</li>
          </ul>
        </div>
      </div>
    );
  };
  
  // Navbar component
  const Navbar = () => {
    return (
      <nav className="bg-gray-900 text-white px-4 py-4 fixed w-full top-0 z-10">
        <div className="container mx-auto flex justify-between items-center">
          <div>
            <span className="font-semibold text-xl">Web Scraper</span>
          </div>
          <div>
            <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
              Profile
            </button>
          </div>
        </div>
      </nav>
    );
  };

const TableWithPagination = ({ data }) => {
  const [currentPage, setCurrentPage] = useState(1);
  const [rowsPerPage] = useState(5); // Change this to control rows per page
  const [scrapedData, setScrapedData] = useState(data);

  // Calculate total pages
  const totalPages = Math.ceil(scrapedData.length / rowsPerPage);

  // Get current page data
  const currentData = scrapedData.slice(
    (currentPage - 1) * rowsPerPage,
    currentPage * rowsPerPage
  );

  // Handle pagination
  const handlePageChange = (pageNumber) => {
    setCurrentPage(pageNumber);
  };

  // CSV download function
  const handleDownloadCSV = () => {
    const parser = new Parser();
    const csv = parser.parse(scrapedData);
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.setAttribute('href', url);
    a.setAttribute('download', 'scraped_data.csv');
    a.click();
  };

  return (
    <div className="overflow-x-auto">
      <div className="min-w-screen min-h-screen bg-gray-100 flex items-center justify-center bg-gray-100 font-sans overflow-hidden">
        <div className="w-full lg:w-5/6">
          <div className="bg-white shadow-md rounded my-6">
            <div className="flex justify-between p-4">
              <h3 className="text-xl font-bold">Scraped Data</h3>
              <button
                onClick={handleDownloadCSV}
                className="bg-blue-500 text-white px-4 py-2 rounded shadow hover:bg-blue-600"
              >
                Download CSV
              </button>
            </div>
            <table className="min-w-max w-full table-auto">
              <thead>
                <tr className="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                  {Object.keys(scrapedData[0] || {}).map((key, index) => (
                    <th key={index} className="py-3 px-6 text-left">{key}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="text-gray-600 text-sm font-light">
                {currentData.map((item, index) => (
                  <tr
                    key={index}
                    className="border-b border-gray-200 hover:bg-gray-100"
                  >
                    {Object.values(item).map((value, subIndex) => (
                      <td key={subIndex} className="py-3 px-6 text-left">
                        {value}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
            <div className="flex justify-center p-4">
              <Pagination
                totalPages={totalPages}
                currentPage={currentPage}
                handlePageChange={handlePageChange}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

// Pagination component
const Pagination = ({ totalPages, currentPage, handlePageChange }) => {
  const pageNumbers = [];

  for (let i = 1; i <= totalPages; i++) {
    pageNumbers.push(i);
  }

  return (
    <nav>
      <ul className="flex space-x-1">
        {pageNumbers.map((number) => (
          <li key={number}>
            <button
              onClick={() => handlePageChange(number)}
              className={`px-4 py-2 rounded ${
                number === currentPage
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-200 text-gray-600'
              }`}
            >
              {number}
            </button>
          </li>
        ))}
      </ul>
    </nav>
  );
};



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




const Home = () => {
    const [url, setUrl] = useState('');
    const [iframeUrl, setIframeUrl] = useState('');
    const [selectors, setSelectors] = useState([]);
    const [iframeRef, setIframeRef] = useState(null); // Ref to the iframe
    const [scrapedData, setScrapedData] = useState([]);

    const handleUrlChange = e => setUrl(e.target.value);

    const handleLoadContent = () => {
      const encodedUrl = encodeURIComponent(url);
      setIframeUrl(`/api/proxy?url=${encodedUrl}`);
    };

    const addSelector = (selector) => {
        if (!selectors.includes(selector)) {
            console.log(selectors, selector);
            setSelectors([...selectors, selector]);
        }
    };

    // const removeSelector = (index) => {
    //     setSelectors(selectors.filter((_, idx) => idx !== index));
    // };

    const removeSelector = (selector, index) => {
        setSelectors(selectors.filter((_, idx) => idx !== index));
        if (iframeRef) {
            // Send a message to the iframe to remove the highlight
            iframeRef.contentWindow.postMessage({
                type: 'REMOVE_HIGHLIGHT',
                selector: selector
            }, '*'); // Use a more restrictive target origin in production
        }
    };

    const fetchScrapedData = async () => {
        if (selectors.length > 0) {
            const fields = selectors.reduce((acc, cur, idx) => {
                if (idx > 0) { // Skip the first selector since it's used as LCA
                    acc[`field${idx}`] = cur.replace(/\.highlight\b\s?/g, '');
                    console.log(cur.replace(/\.highlight\b\s?/g, ''));
                }
                return acc;
            }, {});
    
            try {
                const response = await fetch('http://localhost:8080/scrape', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        url: url,
                        lcaSelector: selectors[0].replace(/\.highlight\b\s?/g, ''), // First selector as LCA
                        groups: {
                            fields: fields
                        }
                    })
                });

                console.log("json: ", JSON.stringify({
                    url: url,
                    lcaSelector: selectors[0], // First selector as LCA
                    groups: {
                        fields: fields
                    }
                }));
    
                if (response.ok) {
                    const data = await response.json();
                    setScrapedData(data);
                    console.log("data:", data);
                } else {
                    console.error("Failed to fetch data:", response.statusText);
                }
            } catch (error) {
                console.error("Error fetching data:", error);
            }
        }
    };
    

    useEffect(() => {
        const handleMessage = event => {
            if (typeof event.data === 'string') {
                addSelector(event.data);
            }
        };

        window.addEventListener('message', handleMessage);
        return () => window.removeEventListener('message', handleMessage);
    }, [selectors]);

    return (
        <div className="container mx-auto p-4 bg-custom-white text-custom-black" style={{ height: '100vh' }}>
            <h1 className="text-3xl font-bold mb-4 bg-pattern-dot">Web Content Embedder</h1>
            <div className="mb-4">
                <input className="shadow appearance-none border rounded py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" 
                       type="text" value={url} onChange={handleUrlChange} placeholder="Enter a URL" />
                <button className="bg-custom-black hover:bg-gray-800 text-custom-white font-bold py-2 px-4 rounded ml-2"
                        onClick={handleLoadContent}>Load Content</button>
            </div>
            <iframe ref={(ref) => setIframeRef(ref)} className="w-full border-gray-300 border" style={{ height: '50vh' }} src={iframeUrl}></iframe>

            <div className="mt-4">
                <h2 className="text-xl font-semibold">Selected CSS Selectors:</h2>
                <ul>
                    {selectors.map((selector, index) => (
                        <li key={index} className="flex justify-between items-center p-2 hover:bg-gray-100">
                            {selector} <button className="bg-custom-black hover:bg-gray-800 text-white font-bold py-1 px-3 rounded"
                                               onClick={() => removeSelector(selector, index)}>Remove</button>
                        </li>
                    ))}
                </ul>
            </div>
            <button className="bg-custom-black hover:bg-gray-800 text-custom-white font-bold py-2 px-4 rounded ml-2"
    onClick={fetchScrapedData}>Fetch Data</button>


    
      <h3 className="text-xl font-semibold">Scraped Data:</h3>
      {/* <TableWithPagination scrapedData={scrapedData} /> */}
      <TableComponent4 scrapedData={scrapedData} />


        </div>
    );
};

export default Home;

