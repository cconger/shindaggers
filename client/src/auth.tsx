import type { Setter, Accessor, Resource } from 'solid-js';
import type { User } from './resources';
import { createSignal, createResource, createEffect } from 'solid-js';

const lsKey = "authtoken";

type authinfo = {
  id: string;
  name: string;
  token: string;
}

const fetchUser = async (token: string | null): Promise<User | null> => {
  if (token === null || token === "") {
    return null;
  }
  let response = await fetch("/api/user/me", {
    headers: {
      'Authorization': token,
    },
  });
  if (response.status === 403 || response.status === 404) {
    // Your token is no good...
    useAuthManager().setToken(null);
    return null
  }
  if (response.status !== 200) {
    throw new Error("unexpected status code " + response.statusText);
  }
  return await response.json().then((res) => res.User);
}

class AuthManager {
  user: Resource<User | null>
  authInfo: authinfo | undefined;
  token: Accessor<string | null>;
  setToken: Setter<string | null>;

  constructor() {
    const rawToken = localStorage.getItem(lsKey);
    [this.token, this.setToken] = createSignal(rawToken);
    [this.user] = createResource(() => this.token() || "", fetchUser);

    createEffect(() => {
      let t = this.token();
      if (t === null) {
        localStorage.removeItem(lsKey);
      } else {
        localStorage.setItem(lsKey, t);
      }
    });
  }
}

let am: AuthManager | undefined;
export const useAuthManager = () => {
  if (am === undefined) {
    am = new AuthManager();
  }
  return am;
}
