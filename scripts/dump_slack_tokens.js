electron = process.mainModule.require('electron');
window = electron.webContents.getAllWebContents()[0];
let config_blob = await window.executeJavaScript('localStorage.localConfig_v2');
let config_obj = JSON.parse(config_blob);
let teams = Object.values(config_obj.teams)
let extracted_teams = [];

teams.forEach(e => {
  extracted_teams.push({
    'name': e.name,
    'token': e.token
  })
});

JSON.stringify(extracted_teams)
