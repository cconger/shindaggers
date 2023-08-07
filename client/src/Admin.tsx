import type { Component, JSX } from 'solid-js';

import { createStore } from 'solid-js/store';


export const AdminPage: Component = () => {

  let fileInputRef = null;

  const [store, setStore] = createStore({
    image: null,
    imagePreview: "",
  })

  let handleImageChange: JSX.EventHandler<HTMLInputElement, Event> = (e) => {
    const files = e.currentTarget.files || [];
    const image = files[0];
    const imagePreview = URL.createObjectURL(image);
    setStore({
      image,
      imagePreview,
    });
  }

  return (
    <div>
      <h1>Admin Page</h1>

      <h2>Functionalty</h2>
      <ul>
        <li>Create new knife</li>
        <li>Modify knives</li>
        <li>Disable knives</li>
        <li>Delete knife</li>
        <li>Issue Knife</li>
        <li>Change Weights</li>
      </ul>

      <div>
        <h2>Image Upload</h2>
        <input type="file" hidden ref={fileInputRef} accept="image/*" onChange={handleImageChange} />
        <div class="button" onClick={() => fileInputRef.click()}>Select Image</div>
        <div class="preview">
          <img src={store.imagePreview} />
        </div>
      </div>
    </div>
  )
}

export const AdminCatalog: Component = () => {
  return (
    <div>AdminCatalog</div>
  )
}

export const AdminKnife: Component = () => {
  return (
    <div>AdminKnife</div>
  )
}
