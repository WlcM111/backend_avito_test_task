import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,           // 10 виртуальных пользователей
  duration: '30s',   // в течение 30 секунд
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // health
  let res = http.get(`${BASE_URL}/health`);
  check(res, {
    'health is 200': (r) => r.status === 200,
  });

  // создаём команду (идемпотентно по team_name)
  res = http.post(`${BASE_URL}/team/add`, JSON.stringify({
    team_name: 'backend',
    members: [
      { user_id: 'u1', username: 'Alice', is_active: true },
      { user_id: 'u2', username: 'Bob',   is_active: true },
      { user_id: 'u3', username: 'Eve',   is_active: true },
    ],
  }), { headers: { 'Content-Type': 'application/json' } });

  check(res, {
    'team add 201 or 400': (r) => r.status === 201 || r.status === 400,
  });

  // создаём PR
  const prId = `pr-${__VU}-${__ITER}`; // уникальные id
  res = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify({
    pull_request_id: prId,
    pull_request_name: 'Load test PR',
    author_id: 'u1',
  }), { headers: { 'Content-Type': 'application/json' } });

  check(res, {
    'create pr 201 or 409': (r) => r.status === 201 || r.status === 409,
  });

  // merge PR
  res = http.post(`${BASE_URL}/pullRequest/merge`, JSON.stringify({
    pull_request_id: prId,
  }), { headers: { 'Content-Type': 'application/json' } });

  check(res, {
    'merge pr 200 or 404': (r) => r.status === 200 || r.status === 404,
  });

  sleep(0.1);
}
