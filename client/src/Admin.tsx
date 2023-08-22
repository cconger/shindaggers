import type { Component, JSX } from 'solid-js';
import { Show, For, Match, Switch, createResource, createSignal } from 'solid-js';
import { A, Navigate, Outlet, useParams } from '@solidjs/router';
import { Rarity, rarities } from './resources';
import type { AdminCollectable, User } from './resources';
import { createStore } from 'solid-js/store';
import { Card } from './Card';
import { useAuthManager } from './LoginButton';
import { Button } from './Button';

import './Admin.css';

type RequireLoginProps = {
  children: JSX.Element;
  fallback?: JSX.Element;
}

export const RequireLogin: Component<RequireLoginProps> = (props) => {
  let am = useAuthManager();

  return (
    <Switch>
      <Match when={am.user.loading}>
        Loading User
      </Match>
      <Match when={am.user.error || !am.user()}>
        <Show when={props.fallback} fallback={<Navigate href="/" />}>
          {props.fallback}
        </Show>
      </Match>
      <Match when={am.user()}>
        {props.children}
      </Match>
    </Switch>
  );
}

export const AdminWrapper: Component = () => {
  return (
    <div class="admin-page">
      <RequireLogin>
        <Outlet />
      </RequireLogin>
    </div>
  );
}

const createCollectable = async (c: AdminCollectable): Promise<AdminCollectable> => {
  let am = useAuthManager();

  let resp = await fetch("/api/admin/collectable", {
    method: "POST",
    body: JSON.stringify({ Collectable: c }),
    headers: {
      "Content-Type": "application/json",
      "Authorization": am.token()!,
    },
  });

  if (!resp.ok) {
    throw new Error("unexpected status code: " + resp.status);
  }

  let body = await resp.json();
  return body.Collectable;
}

export const AdminPage: Component = () => {
  // TODO: Split these into individual pages
  return (
    <>
      <h1>Admin Page</h1>

      <h2>Get Token</h2>
      <Token />

      <h2>New Knife</h2>
      <CollectableForm preview onSubmit={createCollectable} />

      <h2>Events</h2>
      <EventsList />

      <h2>Collectables</h2>
      <CollectableList />
    </>
  )
}


const Token: Component = () => {
  let am = useAuthManager();
  let [show, setShow] = createSignal(false);

  return (
    <div>
      <Show when={show()}>
        {am.token()}
        <div class="button" onClick={() => { setShow(false); }}>Hide</div>
      </Show>
      <Show when={!show()}>
        <div class="button" onClick={() => { window.confirm("Show your token? Do not show on stream!"); setShow(!show()) }} >Show Token</div>
      </Show>
    </div >
  );
}

const EventsList: Component = () => {
  return (
    <div>
      <div class="collectable-table">
        <div class="header">ID</div>
        <div class="header">Name</div>
        <div class="header">Slug</div>
      </div>
    </div>
  )
}

type AdminCollectableList = {
  ApprovalQueue: AdminCollectable[],
  Collectables: AdminCollectable[],
}

const fetchAdminCollectables = async (): Promise<AdminCollectableList> => {
  let am = useAuthManager();

  let resp = await fetch("/api/admin/collectables", {
    headers: {
      "Authorization": am.token()!,
    },
  });

  if (!resp.ok) {
    throw new Error("unexpected status code: " + resp.status)
  }

  let body = await resp.json()

  return body;
}

const CollectableList: Component = () => {
  const [collectables] = createResource(fetchAdminCollectables)

  const [filterDeleted, setFilteredDeleted] = createSignal(true);

  const shown = () => {
    let cs = collectables();
    if (!cs) {
      return [];
    }

    let res = cs.Collectables.slice().reverse();

    if (filterDeleted()) {
      res = res.filter((c) => !c.deleted);
    }

    return res;
  }

  const pending = () => {
    let cs = collectables();
    if (!cs) {
      return [];
    }

    return cs.ApprovalQueue;
  }

  return (
    <div>
      <h3>Pending Approval</h3>
      <div>
        <Switch>
          <Match when={collectables.loading}>
            <div>Loading...</div>
          </Match>
          <Match when={collectables.error}>
            <div>Error: {collectables.error.toString()}</div>
          </Match>
          <Match when={collectables()}>
            <div class="collectable-table">
              <div class="header">ID</div>
              <div class="header">Name</div>
              <div class="header">Author</div>
              <div class="header">Rarity</div>
              <div class="header">Image</div>
              <div class="header">Active</div>
              <For each={pending()}>
                {(collectable) => (
                  <>
                    <div> <A href={`/admin/knife/${collectable.id}`}>{collectable.id}</A> </div>
                    <div> <A href={`/admin/knife/${collectable.id}`}>{collectable.name}</A> </div>
                    <div>{collectable.author.name}</div>
                    <div>{collectable.rarity}</div>
                    <div><A href={collectable.image_url}>{collectable.image_path}</A></div>
                    <div>{collectable.approved ? (collectable.deleted ? "❌" : "✅") : "⚠️"}</div>
                  </>
                )}
              </For>
            </div>
          </Match>
        </Switch>
      </div>
      <h3>Approved</h3>
      <div>
        <label>Show Deleted
          <input
            type="checkbox"
            checked={!filterDeleted()}
            onChange={(e) => setFilteredDeleted(!e.currentTarget.checked)}
          />
        </label>
      </div>
      <Switch>
        <Match when={collectables.loading}>
          <div>Loading...</div>
        </Match>
        <Match when={collectables.error}>
          <div>Error: {collectables.error.toString()}</div>
        </Match>
        <Match when={collectables()}>
          <div class="collectable-table">
            <div class="header">ID</div>
            <div class="header">Name</div>
            <div class="header">Author</div>
            <div class="header">Rarity</div>
            <div class="header">Image</div>
            <div class="header">Active</div>
            <For each={shown()}>
              {(collectable) => (
                <>
                  <div> <A href={`/admin/knife/${collectable.id}`}>{collectable.id}</A> </div>
                  <div> <A href={`/admin/knife/${collectable.id}`}>{collectable.name}</A> </div>
                  <div>{collectable.author.name}</div>
                  <div>{collectable.rarity}</div>
                  <div><A href={collectable.image_url}>{collectable.image_path}</A></div>
                  <div>{collectable.deleted ? "❌" : "✅"}</div>
                </>
              )}
            </For>
          </div>
        </Match>
      </Switch>
    </div>
  )
}


const deleteKnife = async (id: string): Promise<AdminCollectable> => {
  let am = useAuthManager();

  let resp = await fetch("/api/admin/collectable/" + id, {
    method: "DELETE",
    headers: {
      "Authorization": am.token()!,
    },
  });

  if (!resp.ok) {
    throw new Error("unexpected status code: " + resp.status);
  }

  let body = await resp.json();
  return body.Collectable;
}

export type CollectableFormProps = {
  collectable?: AdminCollectable;
  preview?: boolean;
  allowDelete?: boolean;
  authuser?: boolean;
  onSubmit?: (c: AdminCollectable) => Promise<unknown>
}

type ImageUploadResponse = {
  ImagePath: string,
  ImageURL: string,
};

type UploadState = {
  name: string,
  rarity: Rarity,
  author: null | User,
  image: null | File,
  imagePreview: string,
  imagePath: string,
  preview: boolean,
}

export const CollectableForm: Component<CollectableFormProps> = (props) => {
  let fileInputRef: HTMLInputElement | undefined = undefined;

  const [store, setStore] = createStore<UploadState>({
    name: props.collectable?.name || "",
    rarity: props.collectable?.rarity || Rarity.Common,
    author: props.collectable?.author || null,
    image: null,
    imagePreview: props.collectable?.image_url || "",
    imagePath: props.collectable?.image_path || "",
    preview: !!props.collectable,
  })

  let collectable = () => {
    let name = store.name;
    let author = store.author || {
      id: "",
      name: "undefined",
    };

    return {
      id: props.collectable?.id || "",
      name: name,
      author: author,
      rarity: store.rarity,
      image_path: store.imagePath,
      image_url: store.imagePreview,
      deleted: false,
      approved: props.collectable?.approved || true,
    };
  }

  let handleImageChange: JSX.EventHandler<HTMLInputElement, Event> = (e) => {
    const files = e.currentTarget.files;
    if (files === null) { return; }
    const image = files[0];
    const imagePreview = URL.createObjectURL(image);
    setStore({
      image,
      imagePreview,
      preview: true,
    });
  }

  let am = useAuthManager();

  let handleUpload = async (): Promise<ImageUploadResponse> => {

    if (store.image === null) {
      throw new Error("image is null")
    }

    if (!am.token()) {
      throw new Error("not logged in")
    }

    const formdata = new FormData();
    formdata.append("image", store.image);

    let resp = await fetch("/api/image", {
      body: formdata,
      method: 'POST',
      headers: {
        "Authorization": am.token()!,
      },
    })

    if (!resp.ok) {
      throw new Error("unexpected status code:" + resp.status);
    }

    let res: ImageUploadResponse = await resp.json();
    setStore({
      imagePreview: res.ImageURL,
    })
    return res;
  }

  let handleExistingImage: JSX.EventHandler<HTMLInputElement, Event> = (e) => {
    let url = e.currentTarget.value;

    if (store.image) {
      URL.revokeObjectURL(store.imagePreview);
      setStore({ image: null })
    }

    setStore({
      imagePreview: "https://images.shindaggers.io/images/" + url,
      preview: true,
      imagePath: url,
    })
  }

  let handleSubmit = async () => {
    let c = collectable();
    if (store.image !== null) {
      let imageUpload = await handleUpload();

      c.image_path = imageUpload.ImagePath;
      c.image_url = imageUpload.ImageURL;
    }

    if (props.onSubmit) {
      await props.onSubmit(c);
    } else {
      console.error("no handler for submit", c)
    }
  }

  return (
    <div class="form">
      <div class="input-button">
        <input
          placeholder="Name"
          value={store.name}
          onInput={(e) => setStore({ name: e.currentTarget.value })}
        />
      </div>
      <div>
        <select class="selector" id="rarity" name="rarity" onChange={(e) => setStore({ rarity: e.currentTarget.value as Rarity })}>
          <For each={rarities}>
            {(r) => (
              <option value={r} selected={store.rarity == r}>{r.toString()}</option>
            )}
          </For>
        </select>
      </div>
      <Show when={!props.authuser}>
        <UserSearch
          default={props.collectable?.author}
          onUserSelected={(u) => setStore({ author: u })}
        />
      </Show>
      <div class="image-selector">
        <div class="input-button">
          <input
            placeholder="Existing Image"
            value={store.imagePath}
            onChange={handleExistingImage}
          />
        </div>
        <div class="or"><h3>Or</h3></div>
        <div>
          <input type="file" hidden ref={fileInputRef} accept="image/*" onChange={handleImageChange} />
          <div class="button" onClick={() => fileInputRef?.click()}>Upload Image</div>
        </div>
      </div>

      <Show when={store.preview} keyed>
        <div class="controls">
          <Button text="Submit" onClick={handleSubmit} />
          <Show when={props.allowDelete && props.collectable}>
            <Show when={props.collectable?.deleted}><h3>KNIFE IS DELETED</h3></Show>
            <Show when={!props.collectable?.deleted}>
              <Button text="Delete" danger onClick={() => deleteKnife(props.collectable!.id)} />
            </Show>
          </Show>
        </div>
        <h3>Preview</h3>
        <Show when={props.preview}>
          <Card collectable={collectable()} />
        </Show>
      </Show>
    </div>
  )
}

type UserSearchProps = {
  placeholder?: string,
  default?: User,
  onUserSelected(u: User | null): unknown,
}

const searchUsers = async (search: string): Promise<User[]> => {
  if (search == "") {
    return [];
  }

  let resp = await fetch("/api/users?search=" + search)
  if (!resp.ok) {
    throw new Error("unexpected status")
  }

  let users = await resp.json();

  return users.Users;
}

export const UserSearch: Component<UserSearchProps> = (props) => {
  const [search, setSearch] = createSignal("");
  const [valid, setValid] = createSignal(!!props.default);
  const [searchResults] = createResource(() => search(), searchUsers)

  let placeholder = props.placeholder || "User"

  let inputEl: HTMLInputElement | undefined = undefined;
  let selectUser = (user: User) => {
    props.onUserSelected(user);
    setValid(true);
    setSearch("");
    if (inputEl !== undefined) {
      inputEl.value = user.name;
    }
  };

  let cls = () => ({
    "input-button": true,
    "valid": valid(),
    "invalid": !valid(),
  });

  return (
    <>
      <div classList={cls()}>
        <input ref={inputEl} value={props.default?.name || ""} placeholder={placeholder} onInput={(e) => { setValid(false); setSearch(e.target.value); }} />
        <div class="results">
          <Switch>
            <Match when={searchResults.loading}><img src="https://images.shindaggers.io/images/spinner.svg" /></Match>
            <Match when={searchResults.error}><div>{searchResults.error}</div></Match>
            <Match when={searchResults()}>
              <For each={searchResults()}>
                {(user) => (
                  <div onClick={() => selectUser(user)}>{user.name}</div>
                )}
              </For>
            </Match>
          </Switch>
        </div>
      </div>
    </>
  );
}

const fetchAdminCollectable = async (id: string): Promise<AdminCollectable> => {
  let am = useAuthManager();

  let resp = await fetch("/api/admin/collectable/" + id, {
    headers: {
      "Authorization": am.token()!,
    },
  });

  if (!resp.ok) {
    throw new Error("unexpected status code: " + resp.status);
  }

  let body = await resp.json();
  return body.Collectable;
}

const updateCollectable = async (c: AdminCollectable): Promise<AdminCollectable> => {
  let am = useAuthManager();

  if (c.id === "") {
    throw new Error("id is empty cannot update");
  }

  let resp = await fetch("/api/admin/collectable/" + c.id, {
    method: "PUT",
    body: JSON.stringify({ Collectable: c }),
    headers: {
      "Content-Type": "application/json",
      "Authorization": am.token()!,
    },
  });

  if (!resp.ok) {
    throw new Error("unexpected status code: " + resp.status);
  }

  let body = await resp.json();
  return body.Collectable;
}


const approveCollectable = async (id: string): Promise<AdminCollectable> => {
  let am = useAuthManager();

  let resp = await fetch("/api/admin/collectable/" + id + "/approve", {
    method: "POST",
    headers: {
      "Authorization": am.token()!,
    },
  });

  if (!resp.ok) {
    throw new Error("unexpected status code: " + resp.status);
  }

  let body = await resp.json();
  return body.Collectable;
}

export const AdminKnife: Component = () => {
  const params = useParams();

  let [collectable] = createResource(() => params.id, fetchAdminCollectable)
  let [pending, setPending] = createSignal(true);

  let reject = async () => {
    await deleteKnife(params.id)
    setTimeout(() => {
      setPending(false);
    }, 3000);
  };

  let approve = async () => {
    await approveCollectable(params.id)
    setTimeout(() => {
      setPending(false);
    }, 3000);
  };

  return (
    <>
      <h1>AdminKnife</h1>
      <Switch>
        <Match when={collectable.loading}>
          <div>Loading</div>
        </Match>
        <Match when={collectable.error}>
          <div>Error</div>
        </Match>
        <Match when={collectable()}>
          <Show when={!collectable()!.approved && pending()}>
            <Show when={!collectable()!.deleted} fallback={<h3>Knife Rejected</h3>}>
              <div>
                <h3>This knife is pending approval</h3>
                <div class="controls">
                  <Button text="Approve" onClick={approve} />
                  <Button text="Reject" onClick={reject} danger />
                </div>
              </div>
            </Show>
          </Show>
          <h3>Modify Knife:</h3>
          <CollectableForm collectable={collectable()!} onSubmit={updateCollectable} allowDelete preview />
        </Match>
      </Switch>
    </>
  )
}
