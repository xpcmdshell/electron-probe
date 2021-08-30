electron = process.mainModule.require('electron');
JSON.stringify((await electron.session.defaultSession.cookies.get({})))
