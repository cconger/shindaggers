## Usage

```bash
$ npm install # or pnpm install or yarn install
```

## Available Scripts

In the project directory, you can run:

### `npm run dev` or `npm start`

Runs the app in the development mode.<br>
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

The page will reload if you make edits.<br>

By default this will run against the remote API server at https://shindaggers.io

Authentication will not work properly as the server will redirect you to the oauth handler in prod.  If you want to be logged in you must manually copy the localstorage token from your session on https://shindaggers.io to your `localhost`` localstorage.

If you wish to run against a local server you need to override an environment variables.  (I still need to learn vite configuration modes).  If you're running the default go server it will be at http://localhost:8080 so you should set the env var `export SHINDAGGERS_UPSTREAM=https://localhost:8080` before running the vite server.

### `npm run build`

Builds the app for production to the `dist` folder.<br>
It correctly bundles Solid in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.<br>
Your app is ready to be deployed!

## Deployment

### `npm run install`

Builds the app for production and copies to be embedded in the go app
The go app is deployed in the usual way.
