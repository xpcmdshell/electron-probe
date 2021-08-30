electron = process.mainModule.require('electron');
window = electron.webContents.getAllWebContents()[0];
window.loadURL("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
