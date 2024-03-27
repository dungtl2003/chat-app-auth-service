package org.service.auth.chatappauthservice.utils;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.persistence.AttributeConverter;

public class PrivacyConverter implements AttributeConverter<String, String> {

	private static final ObjectMapper mapper = new ObjectMapper();

	@Override
	public String convertToDatabaseColumn(String attribute) {
		try {
			return mapper.writeValueAsString(attribute);
		}
		catch (JsonProcessingException e) {
			return null;
		}
	}

	@Override
	public String convertToEntityAttribute(String dbData) {
		try {
			dbData = dbData != null ? dbData : "{}";
			return mapper.readValue(dbData, String.class);
		}
		catch (JsonProcessingException e) {
			return null;
		}
	}

}
