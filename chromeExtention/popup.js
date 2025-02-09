document.getElementById("start").addEventListener("click", () => {
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      chrome.scripting.executeScript({
        target: { tabId: tabs[0].id },
        files: ["selector-playground.js"],
      });
    });
  });
  
  document.getElementById("stop").addEventListener("click", () => {
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      chrome.scripting.executeScript({
        target: { tabId: tabs[0].id },
        func: () => {
          document.body.style.cursor = "default";
          const overlay = document.getElementById("selector-overlay");
          const ui = document.getElementById("selector-ui");
          if (overlay) overlay.remove();
          if (ui) ui.remove();
        },
      });
    });
  });
  