package org.service.auth.chatappauthservice.utils;

import jakarta.persistence.AttributeConverter;
import jakarta.persistence.Converter;
import org.service.auth.chatappauthservice.entity.enums.Gender;

import java.util.stream.Stream;

public class GenderConverter implements AttributeConverter<Gender, String> {

	@Override
	public String convertToDatabaseColumn(Gender gender) {
		if (gender == null) {
			return null;
		}

		return gender.getName();
	}

	@Override
	public Gender convertToEntityAttribute(String name) {
		if (name == null) {
			return null;
		}

		return Stream.of(Gender.values())
			.filter(gender -> gender.getName().equals(name))
			.findFirst()
			.orElseThrow(IllegalArgumentException::new);
	}

}
