
RECIPE_DIR=${RECIPE_DIR:-../../recipe}

nwsctl \
  apply \
  --recipe-dir ${RECIPE_DIR} \
  --plan ./plan.json \
  "$@"
