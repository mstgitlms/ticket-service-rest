import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 100 },
    { duration: "30s", target: 300 },
    { duration: "30s", target: 500 },
  ],
};

export default function () {
  let payload = JSON.stringify({
    event: "concert",
    stock: 20
  });

  let res = http.post("http://localhost:8080/tickets", payload, {
    headers: { "Content-Type": "application/json" },
  });

  check(res, {
    "status is 200": (r) => r.status === 200,
  });

  if (res.status === 200) {
    let ticket = JSON.parse(res.body);
    http.post(`http://localhost:8080/tickets/${ticket.id}/purchase`);
  }

  sleep(1);
}