// // Storage for recorded actions
// let recordedActions = [];
// let currentWebsite = window.location.origin; // Track the current website

// // Add Overlay to Highlight Elements
// const overlay = document.createElement("div");
// overlay.id = "selector-overlay";
// overlay.style.position = "absolute";
// overlay.style.background = "rgba(0, 128, 255, 0.2)";
// overlay.style.border = "2px solid blue";
// overlay.style.zIndex = "10000";
// overlay.style.pointerEvents = "none"; // Allow clicks to pass through
// document.body.appendChild(overlay);

// document.addEventListener("mousemove", (event) => {
//   const target = event.target;
//   const rect = target.getBoundingClientRect();
//   overlay.style.top = `${rect.top + window.scrollY}px`;
//   overlay.style.left = `${rect.left + window.scrollX}px`;
//   overlay.style.width = `${rect.width}px`;
//   overlay.style.height = `${rect.height}px`;
// });

// // Function to Generate Generalized Selector
// function generateGeneralizedSelector(element) {
//   const tagName = element.tagName.toLowerCase();
//   const classes = Array.from(element.classList).join(".");
//   const selector = classes ? `${tagName}.${classes}` : tagName;

//   // Generate parent hierarchy for more specificity
//   const parent = element.closest("[class]");
//   const parentSelector = parent
//     ? `${parent.tagName.toLowerCase()}.${Array.from(parent.classList).join(".")}`
//     : null;

//   return parentSelector ? `${parentSelector} ${selector}` : selector;
// }

// // Capture Clicks and Record Actions
// document.addEventListener("click", (event) => {
//   const target = event.target;

//   // Prevent recording clicks on the UI itself
//   if (target.closest("#selector-ui")) return;

//   event.preventDefault();
//   event.stopPropagation();

//   const selector = generateGeneralizedSelector(target);

//   // Save the recorded click action
//   recordedActions.push({
//     type: "click",
//     selector,
//     url: window.location.href,
//     timestamp: Date.now(),
//   });

//   // If the clicked element is a link, navigate to the URL
//   if (target.tagName.toLowerCase() === "a" && target.href) {
//     recordedActions.push({
//       type: "navigate",
//       url: target.href,
//       timestamp: Date.now(),
//     });

//     alert(`Recorded Click and Navigation: ${selector}`);
//     window.location.href = target.href;
//   } else {
//     alert(`Recorded Click: ${selector}`);
//   }
// });

// // Detect Page Changes and Reset Actions for New Websites
// window.addEventListener("load", () => {
//   const newWebsite = window.location.origin;
//   if (newWebsite !== currentWebsite) {
//     recordedActions = []; // Clear actions when navigating to a new website
//     currentWebsite = newWebsite; // Update the current website
//   }
// });

// // Add UI for Recording and Downloading Actions
// const ui = document.createElement("div");
// ui.id = "selector-ui";
// ui.style.position = "fixed";
// ui.style.bottom = "10px";
// ui.style.right = "10px";
// ui.style.background = "white";
// ui.style.padding = "10px";
// ui.style.border = "1px solid black";
// ui.style.zIndex = "10001";

// const downloadButton = document.createElement("button");
// downloadButton.innerText = "Download Actions";

// ui.appendChild(downloadButton);
// document.body.appendChild(ui);

// // Download Actions as JSON
// downloadButton.addEventListener("click", () => {
//   if (recordedActions.length === 0) {
//     alert("No actions recorded to download.");
//     return;
//   }

//   const blob = new Blob([JSON.stringify({ website: currentWebsite, actions: recordedActions }, null, 2)], {
//     type: "application/json",
//   });
//   const url = URL.createObjectURL(blob);
//   const a = document.createElement("a");
//   a.href = url;
//   a.download = `${currentWebsite.replace(/https?:\/\//, "").replace(/[\/:]/g, "_")}_actions.json`;
//   document.body.appendChild(a);
//   a.click();
//   document.body.removeChild(a);
// });


// Storage for recorded actions
let recordedActions = [];
let currentWebsite = window.location.origin; // Track the current website

// Add Overlay to Highlight Elements
const overlay = document.createElement("div");
overlay.id = "selector-overlay";
overlay.style.position = "absolute";
overlay.style.background = "rgba(0, 128, 255, 0.2)";
overlay.style.border = "2px solid blue";
overlay.style.zIndex = "10000";
overlay.style.pointerEvents = "none"; // Allow clicks to pass through
document.body.appendChild(overlay);

// Update overlay position on mouse move
document.addEventListener("mousemove", (event) => {
  const target = event.target;
  const rect = target.getBoundingClientRect();
  overlay.style.top = `${rect.top + window.scrollY}px`;
  overlay.style.left = `${rect.left + window.scrollX}px`;
  overlay.style.width = `${rect.width}px`;
  overlay.style.height = `${rect.height}px`;
});

// Function to Generate Generalized Selector
function generateGeneralizedSelector(element) {
  if (!element) return null;

  const tagName = element.tagName.toLowerCase();
  const classes = Array.from(element.classList).join(".");
  const selector = classes ? `${tagName}.${classes}` : tagName;

  // Generate parent hierarchy for more specificity
  const parent = element.closest("[class]");
  const parentSelector = parent
    ? `${parent.tagName.toLowerCase()}.${Array.from(parent.classList).join(".")}`
    : null;

  return parentSelector ? `${parentSelector} ${selector}` : selector;
}

// Capture Clicks and Record Actions
document.addEventListener("click", (event) => {
  let target = event.target;

  // Prevent recording clicks on the UI itself
  if (target.closest("#selector-ui")) return;

  // Handle delegated clicks
  if (!target.matches("a") && target.closest("a")) {
    target = target.closest("a");
  }

  event.preventDefault();
  event.stopPropagation();

  const selector = generateGeneralizedSelector(target);

  // Debugging output
  console.log("Click detected on:", target);
  console.log("Generated selector:", selector);

  // Save the recorded click action
  recordedActions.push({
    type: "click",
    selector,
    url: window.location.href,
    timestamp: Date.now(),
  });

  // Explicit handling for <a> tags with data-run-click-only
  if (target.tagName.toLowerCase() === "a" && target.href) {
    if (target.hasAttribute("data-run-click-only")) {
      console.log("Handling special link with data-run-click-only:", target);

      // Programmatically trigger the click event
      target.click();

      // Save navigation action
      recordedActions.push({
        type: "navigate",
        url: target.href,
        timestamp: Date.now(),
      });
    } else {
      console.log("Navigating to:", target.href);
      recordedActions.push({
        type: "navigate",
        url: target.href,
        timestamp: Date.now(),
      });
      window.location.href = target.href;
    }
  } else {
    alert(`Recorded Click: ${selector}`);
  }
});

// Detect Page Changes and Reset Actions for New Websites
window.addEventListener("load", () => {
  const newWebsite = window.location.origin;
  if (newWebsite !== currentWebsite) {
    recordedActions = []; // Clear actions when navigating to a new website
    currentWebsite = newWebsite; // Update the current website
  }
});

// Add UI for Recording and Downloading Actions
const ui = document.createElement("div");
ui.id = "selector-ui";
ui.style.position = "fixed";
ui.style.bottom = "10px";
ui.style.right = "10px";
ui.style.background = "white";
ui.style.padding = "10px";
ui.style.border = "1px solid black";
ui.style.zIndex = "10001";

const downloadButton = document.createElement("button");
downloadButton.innerText = "Download Actions";

ui.appendChild(downloadButton);
document.body.appendChild(ui);

// Download Actions as JSON
downloadButton.addEventListener("click", () => {
  if (recordedActions.length === 0) {
    alert("No actions recorded to download.");
    return;
  }

  const blob = new Blob([JSON.stringify({ website: currentWebsite, actions: recordedActions }, null, 2)], {
    type: "application/json",
  });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `${currentWebsite.replace(/https?:\/\//, "").replace(/[\/:]/g, "_")}_actions.json`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
});
