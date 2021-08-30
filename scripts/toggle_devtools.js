electron = process.mainModule.require('electron');
window = electron.webContents.getAllWebContents()[0];
window.toggleDevTools()
