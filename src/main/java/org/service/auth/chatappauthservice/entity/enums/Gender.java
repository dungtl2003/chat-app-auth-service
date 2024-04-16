package org.service.auth.chatappauthservice.entity.enums;

import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.SerializerProvider;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.fasterxml.jackson.databind.deser.std.StdDeserializer;
import com.fasterxml.jackson.databind.ser.std.StdSerializer;
import lombok.Getter;

import java.io.IOException;

@Getter
@JsonSerialize(using = GenderSerializer.class)
@JsonDeserialize(using = GenderDeserializer.class)
public enum Gender {

	MALE("male"), FEMALE("female"), OTHER("other"),;

	private final String name;

	Gender(String name) {
		this.name = name;
	}

}

class GenderSerializer extends StdSerializer<Gender> {

	protected GenderSerializer() {
		super(Gender.class);
	}

	@Override
	public void serialize(Gender gender, JsonGenerator jsonGenerator, SerializerProvider serializerProvider)
			throws IOException {
		jsonGenerator.writeStartObject();
		jsonGenerator.writeFieldName("gender");
		jsonGenerator.writeString(gender.getName());
		jsonGenerator.writeEndObject();
	}

}

class GenderDeserializer extends StdDeserializer<Gender> {

	protected GenderDeserializer() {
		super(Gender.class);
	}

	@Override
	public Gender deserialize(JsonParser jsonParser, DeserializationContext deserializationContext) throws IOException {
		String name = jsonParser.getValueAsString();

		for (Gender gender : Gender.values()) {
			if (name.equals(gender.getName())) {
				return gender;
			}
		}

		return null;
	}

}
