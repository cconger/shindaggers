export type User = {
  id: string;
  name: string;
};

export type Collectable = {
  id: string;
  name: string;
  author: User;
  rarity: Rarity;
  image_path: string;
  image_url: string;
};

export type AdminCollectable = Collectable & {
  deleted: boolean;
  approved: boolean;
}

export type IssuedCollectable = Collectable & {
  instance_id: string;
  owner: User;
  verified: boolean;
  subscriber: boolean;
  edition: string;
  issued_at: string;
  deleted: boolean;
};

export enum Rarity {
  Common = 'Common',
  Uncommon = 'Uncommon',
  Rare = 'Rare',
  SuperRare = 'Super Rare',
  UltraRare = 'Ultra Rare',
};

export const rarityclass = (r: Rarity): string => {
  return r.toString().toLowerCase().replaceAll(' ', '-')
};

export const rarities = [
  Rarity.Common,
  Rarity.Uncommon,
  Rarity.Rare,
  Rarity.SuperRare,
  Rarity.UltraRare,
];


const colorMap = {
  [Rarity.Common]: "rgba(38, 184, 0, 0.5)",
  [Rarity.Uncommon]: "rgba(0, 94, 102, 0.5)",
  [Rarity.Rare]: "rgba(81, 1, 101, 0.5)",
  [Rarity.SuperRare]: "rgba(204, 203, 0, 0.5)",
  [Rarity.UltraRare]: "rgba(166, 0, 0, 0.5)",
};

export const rarityColor = (r: Rarity): string => {
  return colorMap[r];
};

export type UserDuelStats = {
  wins: number;
  losses: number;
  ties: number;
}
