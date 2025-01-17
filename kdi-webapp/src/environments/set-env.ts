const setEnv = () => {
    const fs = require('fs');
    const writeFile = fs.writeFile;
    // Configure Angular `environment.ts` file path
    const targetPath = './src/environments/environment.ts';
    // Load node modules
    const colors = require('colors');

    require('dotenv').config({
        path: 'src/environments/.env.local'
    });
    // `environment.ts` file structure
    const envConfigFile = `export const environment = {
    production: true,
    apiUrl: 'https://kdi-web-kuro08-dev.apps.sandbox-m3.1530.p1.openshiftapps.com/api/v1',
    clientId: '${process.env["KDI_WEBAPP_MSAL_CLIENT_ID"]}',
    redirectUri: 'https://kdi-webapp-kuro08-dev.apps.sandbox-m3.1530.p1.openshiftapps.com',
    authority: '${process.env["KDI_WEBAPP_MSAL_AUTHORITY"]}',
    scopes: '${process.env["KDI_WEBAPP_MSAL_SCOPE"]}'?.split(', '),
    };
    `;
    console.log(colors.magenta('The file `environment.ts` will be written with the following content: \n'));
    console.log(colors.magenta('KDI_WEBAPP_WEP_API_ENDPOINT : ', process.env["KDI_WEBAPP_WEP_API_ENDPOINT"]));
    writeFile(targetPath, envConfigFile, (err: any) => {
        if (err) {
            console.error(err);
            throw err;
        } else {
            console.log(colors.magenta(`Angular environment.ts file generated correctly at ${targetPath} \n`));
        }
    });
};

setEnv();
