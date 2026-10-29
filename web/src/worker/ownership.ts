// One scientific computation per tab, independent of the visible route.
let owner: object | undefined;
export function claimScience(candidate: object) {
  if (owner && owner !== candidate) throw new Error("Another scientific computation is active in this tab; cancel it or wait before starting another.");
  owner = candidate;
}
export function releaseScience(candidate: object) { if (owner === candidate) owner = undefined; }
