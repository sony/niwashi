
RECIPE_DIR=${RECIPE_DIR:-../../recipe}

nwsctl init

nwsctl \
  plan \
  --target ./target.yaml \
  --recipe-dir ${RECIPE_DIR} \
  --profile ./profile.yaml \
  --out ./plan.json
