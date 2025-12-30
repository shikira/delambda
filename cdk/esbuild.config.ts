import * as esbuild from 'esbuild';
import { globSync } from 'glob';

const entryPoints = globSync('**/*.ts', {
  ignore: ['node_modules/**', '**/*.d.ts', 'esbuild.config.ts'],
});

const isWatch = process.argv.includes('--watch');

const buildOptions: esbuild.BuildOptions = {
  entryPoints,
  outdir: '.',
  platform: 'node',
  format: 'esm',
  sourcemap: true,
  logLevel: 'info',
};

if (isWatch) {
  const ctx = await esbuild.context(buildOptions);
  await ctx.watch();
  console.log('Watching for changes...');
} else {
  await esbuild.build(buildOptions);
}
