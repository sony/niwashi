
RECIPE_DIR=${RECIPE_DIR:-../../recipe}

nwsctl \
  plan \
  --destroy \
  --recipe-dir ${RECIPE_DIR} \
  --profile ./profile.yaml \
  --out ./destroy-plan.json

nwsctl \
  apply \
  --destroy \
  --recipe-dir ${RECIPE_DIR} \
  --plan ./destroy-plan.json \
  --yes \
  "$@"
