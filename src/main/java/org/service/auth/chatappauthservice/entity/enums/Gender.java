package org.service.auth.chatappauthservice.entity.enums;

import lombok.Getter;

@Getter
public enum Gender {

	MALE("male"), FEMALE("female"), OTHER("other"),;

	private final String name;

	Gender(String name) {
		this.name = name;
	}

}
