import type { Component } from 'solid-js';
import type { Collectable } from '../resources';
import { rarityclass } from '../resources';
import TextArtBG from './TextArtBG';
import './MiniCard.css';


type MiniCardProps = {
  collectable: Collectable;
}

export const MiniCard: Component<MiniCardProps> = (props) => {
  const { collectable } = props;

  const imageURL = "https://images.shindaggers.io/images/" + collectable.image_path;

  const cls = ["mini-card", rarityclass(collectable.rarity)].join(" ");

  return (
    <div class={cls}  >
      <svg class="border border-top" width="130" height="80" viewBox="0 0 272 159" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M136 9H9V159" stroke-width="17" />
        <path d="M160 9H221" stroke-width="17" />
        <path d="M241 9H272" stroke-width="17" />
      </svg>

      <div class="micro-title">
        {collectable.name}
      </div>

      <div class="macro-title">
        <TextArtBG name={collectable.name} lineHeight={50} size={{ width: 200, height: 250 }} />
      </div>

      <div class="card-image">
        <img src={imageURL} draggable="false"/>
      </div>

      <svg class="border border-bottom" width="130" height="80" viewBox="0 0 263 156" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M127 147L254 147L254 1.50204e-05" stroke-width="17" />
        <path d="M110 147H49" stroke-width="17" />
        <path d="M31 147H9.53674e-07" stroke-width="17" />
      </svg>
    </div>
  );
};