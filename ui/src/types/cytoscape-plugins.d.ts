/**
 * Type definitions for Cytoscape plugins
 * These modules lack official TypeScript definitions, so we provide them here
 */

declare module 'cytoscape-dagre' {
  import type cytoscape from 'cytoscape';
  const cytoscapeDagre: cytoscape.Ext;
  export default cytoscapeDagre;
}

declare module 'cytoscape-fcose' {
  import type cytoscape from 'cytoscape';
  const cytoscapeFcose: cytoscape.Ext;
  export default cytoscapeFcose;
}
