// Teste de carga da abertura e listagem de Ordens de Serviço.
//
// Serve para demonstrar o Horizontal Pod Autoscaler: a rampa leva a CPU dos
// pods acima da meta de 70% e o HPA cria réplicas novas. Ver
// docs/KUBERNETES.md, seção "Demonstração do autoscaling".
//
// Uso:
//   make load-test BASE_URL=http://127.0.0.1:30080 K6_EMAIL=admin@... K6_SENHA=...
// ou
//   k6 run -e BASE_URL=... -e K6_EMAIL=... -e K6_SENHA=... test/load/ordens-servico.js
//
// Pré-requisito: banco populado com pelo menos um cliente e um veículo
// (cmd/seed) e um usuário interno para o login.

import http from 'k6/http';
import { check, fail, sleep } from 'k6';

const BASE_URL = (__ENV.BASE_URL || '').replace(/\/$/, '');

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '2m', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1500'],
  },
};

function getFirstId(res, recurso) {
  if (res.status !== 200) {
    fail(`GET ${recurso} retornou ${res.status}: ${res.body}`);
  }
  const items = res.json('items');
  if (!items || items.length === 0) {
    fail(`nenhum registro em ${recurso}; rode o seed antes do teste de carga`);
  }
  return items[0].id;
}

// Roda uma vez antes dos usuários virtuais: faz login e escolhe o cliente e
// o veículo usados em todas as aberturas de OS.
export function setup() {
  if (!BASE_URL || !__ENV.K6_EMAIL || !__ENV.K6_SENHA) {
    fail('defina BASE_URL, K6_EMAIL e K6_SENHA');
  }

  const login = http.post(
    `${BASE_URL}/v1/auth/login`,
    JSON.stringify({ email: __ENV.K6_EMAIL, senha: __ENV.K6_SENHA }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  if (login.status !== 200) {
    fail(`login falhou com status ${login.status}: ${login.body}`);
  }
  const token = login.json('access_token');
  const auth = { headers: { Authorization: `Bearer ${token}` } };

  return {
    token,
    clienteId: getFirstId(http.get(`${BASE_URL}/v1/clientes`, auth), '/v1/clientes'),
    veiculoId: getFirstId(http.get(`${BASE_URL}/v1/veiculos`, auth), '/v1/veiculos'),
  };
}

// 70% das iterações abrem uma OS (escrita) e 30% listam (leitura).
export default function (data) {
  const params = {
    headers: {
      Authorization: `Bearer ${data.token}`,
      'Content-Type': 'application/json',
    },
  };

  if (Math.random() < 0.7) {
    const res = http.post(
      `${BASE_URL}/v1/ordens-servico`,
      JSON.stringify({
        cliente_id: data.clienteId,
        veiculo_id: data.veiculoId,
        quilometragem_entrada: 50000 + Math.floor(Math.random() * 10000),
        observacoes: 'Aberta pelo teste de carga k6',
      }),
      { ...params, tags: { name: 'abrir_os' } },
    );
    check(res, { 'abrir OS: 201': (r) => r.status === 201 });
  } else {
    const res = http.get(`${BASE_URL}/v1/ordens-servico`, { ...params, tags: { name: 'listar_os' } });
    check(res, { 'listar OS: 200': (r) => r.status === 200 });
  }

  sleep(0.5);
}
