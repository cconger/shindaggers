/* @refresh reload */
import { render } from 'solid-js/web';
import { Router, Routes, Route, A } from '@solidjs/router';
import { IfLoggedIn, NavLogin } from './components/LoginButton';
import { AdminWrapper, AdminPage, AdminKnife } from './pages/Admin';

import './admin.css';
import './index.css';

const root = document.getElementById('root');

render(() => (
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
      <Route path="/admin" component={AdminWrapper}>
        <Route path="/" component={AdminPage} />
        <Route path="/knife/:id" component={AdminKnife} />
      </Route>
    </Routes>
  </Router>
), root!);
