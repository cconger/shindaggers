import type { Component, JSX } from 'solid-js';
import { Show, For, Match, Switch, createResource, createSignal } from 'solid-js';
import { A, Navigate, Outlet, useParams } from '@solidjs/router';
import { Rarity, rarities } from '../resources';
import type { AdminCollectable, User } from '../resources';
import { createStore } from 'solid-js/store';
import { Card } from '../components/Card';
import { useAuthManager } from '../auth';
import { UserSearch } from '../components/UserSearch';
import { TextField, Select, MenuItem, Button } from '@suid/material';

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
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Slug</th>
          </tr>
        </thead>
        <tbody>
        </tbody>
      </table>
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
            <table class="collectable-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Name</th>
                  <th>Author</th>
                  <th>Rarity</th>
                  <th>Image</th>
                  <th>Active</th>
                </tr>
              </thead>
              <tbody>
                <For each={pending()}>
                  {(collectable) => (
                    <tr>
                      <td> <A href={`/admin/knife/${collectable.id}`}>{collectable.id}</A></td>
                      <td> <A href={`/admin/knife/${collectable.id}`}>{collectable.name}</A></td>
                      <td>{collectable.author.name}</td>
                      <td>{collectable.rarity}</td>
                      <td><A href={collectable.image_url}>{collectable.image_path}</A></td>
                      <td>{collectable.approved ? (collectable.deleted ? "❌" : "✅") : "⚠️"}</td>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
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
          <table class="collectable-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Name</th>
                <th>Author</th>
                <th>Rarity</th>
                <th>Image</th>
                <th>Active</th>
              </tr>
            </thead>
            <tbody>
              <For each={shown()}>
                {(collectable) => (
                  <tr>
                    <td> <A href={`/admin/knife/${collectable.id}`}>{collectable.id}</A></td>
                    <td> <A href={`/admin/knife/${collectable.id}`}>{collectable.name}</A></td>
                    <td>{collectable.author.name}</td>
                    <td>{collectable.rarity}</td>
                    <td><A href={collectable.image_url}>{collectable.image_path}</A></td>
                    <td>{collectable.approved ? (collectable.deleted ? "❌" : "✅") : "⚠️"}</td>
                  </tr>
                )}
              </For>
            </tbody>
          </table>
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

  let handleExistingImage = (e: any) => {
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
        <TextField
          label="Name"
          variant="outlined"
          value={store.name}
          onChange={(e) => setStore({ name: e.currentTarget.value })}
        />
      </div>
      <div>
        <Select
          label="Rarity"
          variant="outlined"
          value={store.rarity}
          onChange={(e) => setStore({ rarity: e.target.value as Rarity })}
        >
          <MenuItem value={Rarity.Common}>Common</MenuItem>
          <MenuItem value={Rarity.Uncommon}>Uncommon</MenuItem>
          <MenuItem value={Rarity.Rare}>Rare</MenuItem>
          <MenuItem value={Rarity.SuperRare}>Super Rare</MenuItem>
          <MenuItem value={Rarity.UltraRare}>Ultra Rare</MenuItem>
        </Select>
      </div>
      <Show when={!props.authuser}>
        <UserSearch
          default={props.collectable?.author}
          onUserSelected={(u) => setStore({ author: u })}
        />
      </Show>
      <div class="image-selector">
        <div class="input-button">
          <TextField
            label="Image Path"
            value={store.imagePath}
            onChange={handleExistingImage}
          />
        </div>
        <div class="or"><h3>Or</h3></div>
        <div>
          <input type="file" hidden ref={fileInputRef} accept="image/*" onChange={handleImageChange} />
          <Button variant="contained" size="large" onClick={() => fileInputRef?.click()}>Upload Image</Button>
        </div>
      </div>

      <Show when={store.preview} keyed>
        <div class="controls">
          <Button variant="contained" size="large" onClick={handleSubmit}>Submit</Button>
          <Show when={props.allowDelete && props.collectable}>
            <Show when={props.collectable?.deleted}><h3>KNIFE IS DELETED</h3></Show>
            <Show when={!props.collectable?.deleted}>
              <Button variant="contained" color="error" onClick={() => deleteKnife(props.collectable!.id)}>Delete</Button>
            </Show>
          </Show>
        </div>
        <h3>Preview</h3>
        <Show when={props.preview}>
          <Card collectable={collectable()} />
        </Show>
      </Show>
    </div >
  )
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
                  <Button variant="contained" color="success" size="large" onClick={approve}>Approve</Button>
                  <Button variant="contained" color="error" size="large" onClick={reject}>Reject</Button>
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
