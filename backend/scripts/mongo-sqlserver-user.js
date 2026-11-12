const technical = db.getSiblingDB("elbravodeljac_tecnico_dev");
technical.createUser({
  user: "jac_app",
  pwd: process.env.JAC_MONGO_APP_PASSWORD,
  roles: [{ role: "readWrite", db: technical.getName() }],
});
