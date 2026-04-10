export type TreeNodeWithKey<T extends { children?: T[]; id: string }> = Omit<
  T,
  'children'
> & {
  children?: Array<TreeNodeWithKey<T>>;
  key: string;
};

export function withTreeKeys<T extends { children?: T[]; id: string }>(
  nodes: T[],
): Array<TreeNodeWithKey<T>> {
  return nodes.map((node) => ({
    ...node,
    key: node.id,
    children: node.children?.length ? withTreeKeys(node.children) : undefined,
  }));
}
