	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
count:		.word	0
str:		.asciiz "Number of digits in n: "

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime

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
	lw	$3, n		# n -> $3
	ble	$3, 0, exit

	lw	$3, n		# n -> $3
	div	$3, $3, 10
	sw	$3, n		# spilled n, freed $3
	lw	$3, count	# count -> $3
	addi	$3, $3, 1
	# Store dirty variables back into memory
	sw	$3, count
	j	while

exit:
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 1
	lw	$3, count	# count -> $3
	move	$4, $3
	syscall
	li	$2, 10
	syscall
	.end main
