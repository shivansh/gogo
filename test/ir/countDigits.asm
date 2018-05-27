	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
count:		.word	0
str:		.asciiz "Number of digits in n: "

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
	li	$3, 0		# count -> $3
	# Store dirty variables back into memory
	sw	$3, count

while:
	lw	$3, n
	ble	$3, 0, exit
	# Store dirty variables back into memory

	lw	$3, n
	div	$3, $3, 10	# n -> $3
	sw	$3, n		# spilled n, freed $3
	lw	$3, count
	addi	$3, $3, 1	# count -> $3
	# Store dirty variables back into memory
	sw	$3, count
	j	while

exit:
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 1
	lw	$3, count
	move	$4, $3
	syscall
	# Store dirty variables back into memory
	li	$2, 10
	syscall
	.end main
