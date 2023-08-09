/* @refresh reload */
import { render } from 'solid-js/web';
import { Router, Routes, Route, A } from '@solidjs/router';

import './index.css';
import { Catalog } from './Catalog';
import { CatalogCard } from './CatalogCard';
import { Home } from './Home';
import { Pull } from './Pull';
import { NavLogin, LoginLander } from './LoginButton';
import { UserCollection } from './User';
import { ButtonTest } from './Button';
import { AdminPage, AdminKnife } from './Admin';

const root = document.getElementById('root');

if (import.meta.env.DEV && !(root instanceof HTMLElement)) {
  throw new Error(
    'Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got misspelled?',
  );
}

render(() => (
  <Router>
    <nav class="masternav">
      <nav class="pages">
        <A href="/">Shindaggers</A>
        <A href="/catalog">Catalog</A>
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
      <Route path="/admin" component={AdminPage} />
      <Route path="/admin/knife/:id" component={AdminKnife} />
      <Route path="/buttontest" component={ButtonTest} />
    </Routes>
  </Router>
), root!);
