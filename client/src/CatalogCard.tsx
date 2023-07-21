import type { Component } from 'solid-js';
import { createResource, Switch, Match } from 'solid-js';
import { useParams } from '@solidjs/router';
import type { Collectable } from './resources';
import { Card } from './Card';

const fetchCollectable = async (id: string): Promise<Collectable> => {
  let response = await fetch("/api/collectable/" + id)
  if (response.status === 404) {
    throw new Error("Collectable does not exist")
  }
  return await response.json().then((resp) => {
    if (resp.Collectable) return resp.Collectable;
    throw new Error("Unexpected Datatype")
  })
}

export const CatalogCard: Component = (props) => {
  const params = useParams();

  const [collectable] = createResource(() => params.id, fetchCollectable);

  return (
    <div>
      <Switch>
        <Match when={collectable.loading}>
          <div>Loading</div>
        </Match>
        <Match when={collectable.error}>
          <div>{collectable.error.toString()}</div>
        </Match>
        <Match when={collectable()}>
          <Card collectable={collectable()!} />
        </Match>
      </Switch>
    </div>
  );
};

export default CatalogCard;
