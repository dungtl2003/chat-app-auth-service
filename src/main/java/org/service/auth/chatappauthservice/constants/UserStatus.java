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
@JsonSerialize(using = UserStatusSerializer.class)
@JsonDeserialize(using = UserStatusDeserializer.class)
public enum UserStatus {

	ONLINE("ONLINE"), OFFLINE("OFFLINE"),;

	private final String name;

	UserStatus(String name) {
		this.name = name;
	}

}

class UserStatusSerializer extends StdSerializer<UserStatus> {

	protected UserStatusSerializer() {
		super(UserStatus.class);
	}

	@Override
	public void serialize(UserStatus status, JsonGenerator jsonGenerator, SerializerProvider serializerProvider)
			throws IOException {
		jsonGenerator.writeStartObject();
		jsonGenerator.writeFieldName("status");
		jsonGenerator.writeString(status.getName());
		jsonGenerator.writeEndObject();
	}

}

class UserStatusDeserializer extends StdDeserializer<UserStatus> {

	protected UserStatusDeserializer() {
		super(UserStatus.class);
	}

	@Override
	public UserStatus deserialize(JsonParser jsonParser, DeserializationContext deserializationContext)
			throws IOException {
		String name = jsonParser.getValueAsString();

		for (UserStatus status : UserStatus.values()) {
			if (name.equals(status.getName())) {
				return status;
			}
		}

		return null;
	}

}
