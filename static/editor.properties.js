var properties = window.PROPERTIES;

for (var i in properties) {
  var property = properties[i];
  Survey.JsonObject.metaData.addProperty("questionbase", {
    name: property.key + ":" + property.type,
    choices: property.choices
  });
}
