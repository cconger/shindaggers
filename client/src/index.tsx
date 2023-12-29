/* @refresh reload */
import { render } from 'solid-js/web';
import { Router, Routes, Route, A } from '@solidjs/router';
import { lazy } from 'solid-js';
import { ThemeProvider, createTheme } from '@suid/material';

import './index.css';
import { Catalog } from './pages/Catalog';
import { CatalogCard } from './pages/CatalogCard';
import { Home } from './pages/Home';
import { Pull } from './pages/Pull';
import { IfLoggedIn, NavLogin, LoginLander } from './components/LoginButton';
import { UserCollection } from './pages/User';
import { UITest } from './pages/UITest';
import { FourOhFour } from './pages/FourOhFour';
import { Creator } from './pages/Creator';
import { Event } from './pages/Event';

const AdminWrapper = lazy(async () => {
  let admin = await import('./pages/Admin')
  return { default: admin.AdminWrapper };
});

const AdminPage = lazy(async () => {
  let admin = await import('./pages/Admin')
  return { default: admin.AdminPage };
});

const AdminKnife = lazy(async () => {
  let admin = await import('./pages/Admin')
  return { default: admin.AdminKnife };
});

const root = document.getElementById('root');

if (import.meta.env.DEV && !(root instanceof HTMLElement)) {
  throw new Error(
    'Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got misspelled?',
  );
}

const theme = createTheme({
  palette: {
    mode: 'dark',
  },
})

render(() => (
  <ThemeProvider theme={theme}>
    <Router>
      <nav class="masternav">
        <nav class="pages">
          <A href="/">Shindaggers</A>
          <A href="/catalog">Catalog</A>
          <IfLoggedIn>
            <A href="/creator">Create</A>
          </IfLoggedIn>
        </nav>
        <nav class="panel">
          <NavLogin />
        </nav>
      </nav>
      <Routes>
        <Route path="/" component={Home} />
        <Route path="/login" component={LoginLander} />

        <Route path="/knife/:id" component={Pull} />
        <Route path="/user/:id" component={UserCollection} />
        <Route path="/catalog" component={Catalog} />
        <Route path="/catalog/:id" component={CatalogCard} />
        <Route path="/creator" component={Creator} />
        <Route path="/event/:slug" component={Event} />
        <Route path="/uitest" component={UITest} />
        <Route path="/admin" component={AdminWrapper}>
          <Route path="/" component={AdminPage} />
          <Route path="/knife/:id" component={AdminKnife} />
        </Route>
        <Route path="*" component={FourOhFour} />
      </Routes>
    </Router>
  </ThemeProvider>
), root!);
