$.get(
  "properties.json?access_token=" +
    $.QueryString.access_token +
    "&survey=" +
    $.QueryString.survey,
  function(properties) {
    for (var i in properties) {
      var property = properties[i];
      Survey.JsonObject.metaData.addProperty("questionbase", {
        name: property.key + ":" + property.type,
        choices: property.choices
      });
    }
  }
);
