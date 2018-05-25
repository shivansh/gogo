# Test to find floor value of logarithm (base 2 and base 10) of a number

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
base2Str:	.asciiz "log2(n): "
base10Str:	.asciiz "\nlog10(n): "

	.text


	.globl main
	.ent main
main:
	li	$2, 4
	la	$4, nStr
	syscall
	li	$2, 5
	syscall
	move	$3, $2
	sw	$3, n		# spilled n, freed $3
	li	$3, -1		# i -> $3
	# Store variables back into memory
	sw	$3, i

while:
	lw	$3, n
	srl	$3, $3, 1	# n -> $3
	lw	$5, i
	addi	$5, $5, 1	# i -> $5
	# Store variables back into memory
	sw	$3, n
	sw	$5, i
	bgt	$3, 0, while

	li	$2, 4
	la	$4, base2Str
	syscall
	li	$2, 1
	lw	$3, i
	move	$4, $3
	syscall
	li	$2, 4
	la	$4, base10Str
	syscall
	# Ideally
	div	$3, $3, 3	# i -> $3
	li	$2, 1
	move	$4, $3
	syscall
	# Store variables back into memory
	sw	$3, i
	li	$2, 10
	syscall
	.end main
