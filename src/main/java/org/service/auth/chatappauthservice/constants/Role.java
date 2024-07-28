package org.service.auth.chatappauthservice.constants;

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
@JsonSerialize(using = RoleSerializer.class)
@JsonDeserialize(using = RoleDeserializer.class)
public enum Role {

	ADMIN("ADMIN"), USER("USER"),;

	private final String name;

	Role(String name) {
		this.name = name;
	}

}

class RoleSerializer extends StdSerializer<Role> {

	protected RoleSerializer() {
		super(Role.class);
	}

	@Override
	public void serialize(Role role, JsonGenerator jsonGenerator, SerializerProvider serializerProvider)
			throws IOException {
		jsonGenerator.writeStartObject();
		jsonGenerator.writeFieldName("role");
		jsonGenerator.writeString(role.getName());
		jsonGenerator.writeEndObject();
	}

}

class RoleDeserializer extends StdDeserializer<Role> {

	protected RoleDeserializer() {
		super(Role.class);
	}

	@Override
	public Role deserialize(JsonParser jsonParser, DeserializationContext deserializationContext) throws IOException {
		String name = jsonParser.getValueAsString();

		for (Role role : Role.values()) {
			if (name.equals(role.getName())) {
				return role;
			}
		}

		return null;
	}

}