import * as fs from 'fs';
import * as path from 'path';
import fastify from 'fastify';

const server = fastify({ logger: true });
const indexHtml = fs.readFileSync(path.resolve(__dirname, '../../public/index.html'));

server.get('/', async (req, reply) => {
  reply
    .type('text/html')
    .send(indexHtml);
});

server.get('/script/*', async (req, reply) => {
  let filename = req.url.substr(8);
  if (!filename.endsWith('.js')) {
    filename += '.js';
  }

  try {
    const data = await fs.promises.readFile(path.resolve(__dirname, '../../dist', filename), 'utf8');
    reply
      .type('text/javascript')
      .send(data);
  } catch (err) {
    req.log.error({ filename }, 'script not found');
    reply.status(404).send();
  }
});

const start = async () => {
  try {
    await server.listen(3000);
  } catch (err) {
    server.log.error(err);
    process.exit(1);
  }
};
start();
