# Test to declare and print a string

	.data
str:		.asciiz "Hello World!"

	.text


	.globl main
	.ent main
main:
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 10
	syscall
	.end main
