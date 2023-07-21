import type { Component } from 'solid-js';
import { onMount, Show } from 'solid-js';
import type { IssuedCollectable, Collectable } from './resources';
import { rarityclass } from './resources';
import TextArtBG from './TextArtBG';
import './Card.css';
import VanillaTilt from 'vanilla-tilt';

type CardProps = {
  collectable: Collectable;
  issuedCollectable?: IssuedCollectable
}

export const Card: Component<CardProps> = (props) => {

  const { collectable, issuedCollectable } = props;
  const imageURL = "https://images.shindaggers.io/images/" + collectable.image_path;
  const cls = ["card", rarityclass(collectable.rarity)].join(" ");

  let card: HTMLDivElement | undefined;
  onMount(() => {
    if (card !== undefined) {
      // TODO: handle this inside framework
      VanillaTilt.init(card)
    }
  })

  return (
    <div class={cls} ref={card}>
      <svg class="border border-top" width="260" height="160" viewBox="0 0 272 159" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M136 9H9V159" stroke-width="17" />
        <path d="M160 9H221" stroke-width="17" />
        <path d="M241 9H272" stroke-width="17" />
      </svg>

      <div class="card-label">
        <div class="justified">
          <div>{collectable.rarity}</div>
        </div>
        <div class="edition">
        </div>
      </div>

      <div class="micro-title">
        {collectable.name}
      </div>

      <div class="macro-title">
        <TextArtBG name={collectable.name} lineHeight={120} size={{ width: 480, height: 500 }} fontSize={100} />
      </div>

      <div class="card-image">
        <img src={imageURL} />
      </div>

      <div class="badges">
        <Show when={issuedCollectable?.deleted}>
          <div class="deleted" title="This knife was deleted. 🤫"></div>
        </Show>
        <Show when={issuedCollectable?.verified}>
          <div class="verified" title="Verified Issue"></div>
        </Show>
        <Show when={issuedCollectable?.subscriber}>
          <div class="subscribed" title="Subscriber Issue"></div>
        </Show>
      </div>

      <div class="card-attribution">
        <div class="label">Crafted By</div>
        <div>{collectable.author.name}</div>
      </div>

      <svg class="border border-bottom" width="260" height="160" viewBox="0 0 263 156" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M127 147L254 147L254 1.50204e-05" stroke-width="17" />
        <path d="M110 147H49" stroke-width="17" />
        <path d="M31 147H9.53674e-07" stroke-width="17" />
      </svg>
    </div>
  )
};


